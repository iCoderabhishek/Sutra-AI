import asyncio
import json
from typing import Callable
import openai

from libs.openrouter_client import generate_content_openai, get_free_models, AGENT_TEMPERATURE
from libs.env import settings
from agent.tools import get_openai_tool_declarations, dispatch_tool, resolve_allowed
from agent.prompts import resolve_system_prompt
from agent.guardrails import (
    InputRejected,
    check_input,
    fallback_report,
    parse_report,
    sanitize_report,
    tool_preview,
    wrap_tool_result,
)
from utils.cost_tracker import CostTracker

MAX_RETRIES = 3
# How many times the model may fix a final reply that isn't valid report JSON.
MAX_FORMAT_REPAIRS = 2

async def _call_with_fallback(kwargs, selected_model: str | None = None):
    """
    Calls OpenRouter with fallback. If a model fails, it tries the next one in the FREE_MODELS list.
    """
    models_to_try = get_free_models().copy()
    if selected_model and selected_model != "auto":
        if selected_model in models_to_try:
            models_to_try.remove(selected_model)
        models_to_try.insert(0, selected_model)
        
    last_error = None
    for attempt in range(len(models_to_try)):
        model = models_to_try[attempt]
        kwargs["model"] = model
        try:
            response = await generate_content_openai(**kwargs)
            if not getattr(response, "choices", None):
                raise Exception("Empty response or no choices returned")
            return response
        except Exception as e:
            last_error = e
            print(f"[loop] Model {model} failed: {e}. Switching to next model...")
            await asyncio.sleep(1) # Small backoff before switching
            
    raise Exception(f"All fallback models failed. Last error: {last_error}")

MAX_ITERATIONS = 10

async def run_agent(
    goal: str,
    stream_callback: Callable,
    tools: list[str] | None = None,
    system_prompt: str | None = None,
    template: str | None = None,
    instruction: str | None = None,
    email: str | None = None,
    model: str | None = None,
) -> str:
    """
    Core ReAct agent loop using OpenRouter.
    Runs up to MAX_ITERATIONS rounds of: Think → Act (tool call) → Observe.
    """
    try:
        goal, instruction = check_input(goal, instruction)
    except InputRejected as e:
        await stream_callback({"step": "Input blocked", "status": "error", "content": str(e)})
        return f"Input blocked: {e}"

    allowed = resolve_allowed(tools, email)
    declarations = get_openai_tool_declarations(allowed)

    task = goal if not instruction else f"{goal}\n\nAdditional instructions:\n{instruction}"

    messages = []
    
    sys_prompt = resolve_system_prompt(template, system_prompt, email, allowed)
    if sys_prompt:
        messages.append({"role": "system", "content": sys_prompt})
        
    messages.append({"role": "user", "content": task})

    kwargs = {
        "messages": messages,
        "temperature": AGENT_TEMPERATURE,
    }
    if declarations:
        kwargs["tools"] = declarations
        kwargs["tool_choice"] = "auto"

    initial_model = model if model and model != "auto" else get_free_models()[0]
    cost = CostTracker(model=initial_model)
    repairs = 0

    try:
        for i in range(MAX_ITERATIONS):
            response = await _call_with_fallback(kwargs, selected_model=model)
            cost.model = response.model
            cost.add(response)

            spent = cost.total()["cost_usd"]
            if spent >= settings.MAX_RUN_COST_USD:
                await stream_callback({
                    "step": "Cost ceiling reached",
                    "status": "error",
                    "content": f"Run stopped after spending ${spent:.4f}, which reached the ${settings.MAX_RUN_COST_USD:.2f} per-run limit.",
                    "iteration": i + 1,
                    "cost": cost.total(),
                })
                return f"Agent stopped: per-run cost limit of ${settings.MAX_RUN_COST_USD:.2f} reached."

            output_message = response.choices[0].message
            
            # Need to convert Message object to dict for appending to messages
            msg_dict = {"role": "assistant"}
            if output_message.content:
                msg_dict["content"] = output_message.content
            else:
                msg_dict["content"] = ""
                
            if output_message.tool_calls:
                msg_dict["tool_calls"] = []
                for tc in output_message.tool_calls:
                    msg_dict["tool_calls"].append({
                        "id": tc.id,
                        "type": "function",
                        "function": {
                            "name": tc.function.name,
                            "arguments": tc.function.arguments
                        }
                    })
                    
            messages.append(msg_dict)

            tool_calls = output_message.tool_calls
            
            if not tool_calls:
                raw = output_message.content or ""
                report, problem = parse_report(raw)

                # Output guardrail: the final answer must be a valid report.
                # Ask the model to fix it before falling back.
                if report is None and repairs < MAX_FORMAT_REPAIRS:
                    repairs += 1
                    await stream_callback({
                        "step": "Formatting report",
                        "status": "running",
                        "iteration": i + 1,
                        "cost": cost.total(),
                    })
                    messages.append({
                        "role": "user",
                        "content": (
                            f"Your reply did not match the required format ({problem}). "
                            "Reply again with only the JSON object described in OUTPUT FORMAT."
                        ),
                    })
                    continue

                report = sanitize_report(report or fallback_report(raw))
                await stream_callback({
                    "step": "Final Answer",
                    "status": "done",
                    "content": report.model_dump_json(),
                    "iteration": i + 1,
                    "cost": cost.total(),
                })
                return report.summary

            for call in tool_calls:
                tool_name = call.function.name
                try:
                    tool_args = json.loads(call.function.arguments)
                except:
                    tool_args = {}
                tool_id = call.id
                
                await stream_callback({
                    "step": f"Running {tool_name}",
                    "status": "running",
                    "args": tool_args,
                    "iteration": i + 1,
                    "cost": cost.total(),
                })

                result = await dispatch_tool(tool_name, tool_args, allowed)

                await stream_callback({
                    "step": f"Completed {tool_name}",
                    "status": "done",
                    "result_preview": tool_preview(str(result)),
                    "iteration": i + 1,
                    "cost": cost.total(),
                })

                messages.append(
                    {
                        "role": "tool", 
                        "tool_call_id": tool_id, 
                        "content": wrap_tool_result(tool_name, str(result))
                    }
                )

        await stream_callback({
            "step": "Max iterations reached",
            "status": "error",
            "iteration": MAX_ITERATIONS,
            "cost": cost.total(),
        })
        return "Agent reached the maximum number of iterations without completing the task."

    except Exception as e:
        await stream_callback({
            "step": "Agent Error",
            "status": "error",
            "content": str(e),
            "cost": cost.total(),
        })
        return f"Agent encountered an error: {e}"
