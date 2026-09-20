from typing import Callable

from google.genai import types

from libs.gemini_client import gemini_client, MODEL_ID, AGENT_TEMPERATURE
from libs.env import settings
from agent.tools import get_gemini_tool_declarations, dispatch_tool

MAX_ITERATIONS = 10
# may change later based on the agent performance

async def run_agent(goal: str, stream_callback: Callable) -> str:
    """
    Core ReAct agent loop.
    Runs up to MAX_ITERATIONS rounds of: Think → Act (tool call) → Observe.
    Streams trace events via stream_callback for SSE.

    Args:
        goal:            The user's task description.
        stream_callback: An async callable that receives trace event dicts.

    Returns:
        The agent's final answer as a string.
    """

    #  Conversation history (grows each iteration)
    # instead of init a langgraph here we are init a array of types.Content
    # and append to it in each iteration as context
    
    contents: list[types.Content] = [
        types.Content(
            role="user",
            parts=[types.Part(text=goal)]
        )
    ]

    # Gemini call config (constant across all iterations)
    config = types.GenerateContentConfig(
        system_instruction=settings.SYSTEM_PROMPT,
        tools=[get_gemini_tool_declarations()],
        temperature=AGENT_TEMPERATURE,
    )
