import asyncio
from typing import Callable

from google.genai import types
from google.genai import errors as genai_errors

from libs.gemini_client import gemini_client, MODEL_ID, AGENT_TEMPERATURE
from libs.env import settings
from agent.tools import get_gemini_tool_declarations, dispatch_tool, resolve_allowed
from agent.prompts import resolve_system_prompt
from utils.cost_tracker import CostTracker

MAX_RETRIES = 3

async def _call_with_retry(contents, config):
    """
    Calls Gemini with exponential backoff on rate-limit (429) and
    service-unavailable (503) errors. Raises on the final attempt.
    """
    for attempt in range(MAX_RETRIES):
        try:
            return await gemini_client.aio.models.generate_content(
                model=MODEL_ID,
                contents=contents,
                config=config,
            )
        except (genai_errors.ClientError, genai_errors.ServerError) as e:
            # Only retry on 429 (rate limit) or 503 (service unavailable)
            # Let other 4xx errors (auth, bad request) fail fast
            if getattr(e, "status_code", None) not in (429, 503):
                raise
            if attempt == MAX_RETRIES - 1:
                raise  # exhausted all retries — let the outer handler catch it
            wait = 2 ** (attempt + 1)  # 2s → 4s → 8s
            print(f"[loop] Rate limited ({e.status_code}) — attempt {attempt + 1}/{MAX_RETRIES}, retrying in {wait}s...")
            await asyncio.sleep(wait)


MAX_ITERATIONS = 10
# MAX_RETRIES = 3 is defined above _call_with_retry
# may change later based on the agent performance

async def run_agent(
    goal: str,
    stream_callback: Callable,
    tools: list[str] | None = None,
    system_prompt: str | None = None,
    template: str | None = None,
    instruction: str | None = None,
    email: str | None = None,
) -> str:
    """
    Core ReAct agent loop.
    Runs up to MAX_ITERATIONS rounds of: Think → Act (tool call) → Observe.
    Streams trace events via stream_callback for SSE.
    """

    allowed = resolve_allowed(tools, email)
    declarations = get_gemini_tool_declarations(allowed)

    #  Conversation history (grows each iteration)
    # instead of init a langgraph here we are init a array of types.Content
    # and append to it in each iteration as context

    task = goal if not instruction else f"{goal}\n\nAdditional instructions:\n{instruction}"

    contents: list[types.Content] = [
        types.Content(
            role="user",
            parts=[types.Part(text=task)]
        )
    ]

    # Gemini call config (constant across all iterations)
    config = types.GenerateContentConfig(
        system_instruction=resolve_system_prompt(template, system_prompt, email, allowed),
        temperature=AGENT_TEMPERATURE,
        # Omitted when nothing is allowed: Gemini rejects an empty tool list.
        **({"tools": [declarations]} if declarations else {}),
    )

    cost = CostTracker(model=MODEL_ID)

    try:
        for i in range(MAX_ITERATIONS):
            response = await _call_with_retry(contents, config)
            cost.add(response)  # accumulate tokens after every LLM call

            spent = cost.total()["cost_usd"]
            if spent >= settings.MAX_RUN_COST_USD:
                await stream_callback({
                    "step": "Cost ceiling reached",
                    "status": "error",
                    "content": (
                        f"Run stopped after spending ${spent:.4f}, which reached the "
                        f"${settings.MAX_RUN_COST_USD:.2f} per-run limit."
                    ),
                    "iteration": i + 1,
                    "cost": cost.total(),
                })
                return (
                    f"Agent stopped: per-run cost limit of "
                    f"${settings.MAX_RUN_COST_USD:.2f} reached."
                )

            res = response.candidates[0].content

            if not response.function_calls:
                final_text = response.text
                await stream_callback({
                    "step": "Final Answer",
                    "status": "done",
                    "content": final_text,
                    "iteration": i + 1,
                    "cost": cost.total(),
                })
                return final_text

            contents.append(res)

            # Collect all function response parts for this round
            function_response_parts: list[types.Part] = []

            for call in response.function_calls:
                await stream_callback({
                    "step": f"Running {call.name}",
                    "status": "running",
                    "args": dict(call.args),
                    "iteration": i + 1,
                    "cost": cost.total(),
                })

                result = await dispatch_tool(call.name, dict(call.args), allowed)

                await stream_callback({
                    "step": f"Completed {call.name}",
                    "status": "done",
                    "result_preview": result[:200],
                    "iteration": i + 1,
                    "cost": cost.total(),
                })

                function_response_parts.append(
                    types.Part.from_function_response(
                        name=call.name,
                        response={"result": result},
                    )
                )

            # Append all tool results as a single "user" turn
            contents.append(
                types.Content(role="user", parts=function_response_parts)
            )

        # Max iterations reached
        await stream_callback({
            "step": "Max iterations reached",
            "status": "error",
            "iteration": MAX_ITERATIONS,
            "cost": cost.total(),
        })
        return "Agent reached the maximum number of iterations without completing the task."

    # Top-level error handler
    except Exception as e:
        await stream_callback({
            "step": "Agent Error",
            "status": "error",
            "content": str(e),
            "cost": cost.total(),
        })
        return f"Agent encountered an error: {e}"
