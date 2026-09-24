from google.genai import types
from pydantic import ValidationError

from tools.scrapper import ScrapperTool
from tools.search import WebSearchTool
from tools.emailer import EmailTool

TOOLS = [
    WebSearchTool(),
    ScrapperTool(),
    EmailTool(),
]

# name → instance lookup
TOOL_MAP = {t.name: t for t in TOOLS}

ALL_TOOL_NAMES = frozenset(TOOL_MAP)


def resolve_allowed(
    requested: list[str] | None,
    email: str | None = None,
) -> frozenset[str]:
    """
    Resolve the tool names this run may call.

    None means all tools, [] means none, and unknown names are dropped rather
    than raising. send_email is added whenever an email is set, since a run with
    a delivery address that cannot send would fail silently.
    """
    if requested is None:
        allowed = set(ALL_TOOL_NAMES)
    else:
        allowed = {name for name in requested if name in TOOL_MAP}

    if email:
        allowed.add(EmailTool.name)

    return frozenset(allowed)


def _clean_schema(schema: dict) -> dict:
    """Strip 'title' fields that Pydantic adds but Gemini rejects."""
    schema.pop("title", None)
    for prop in schema.get("properties", {}).values():
        prop.pop("title", None)
    return schema


def get_gemini_tool_declarations(allowed: frozenset[str] | None = None) -> types.Tool | None:
    """
    Convert the allowed tools into a single Gemini types.Tool.

    Returns None when nothing is allowed; Gemini rejects a Tool with an empty
    function_declarations list, so the caller must omit `tools` entirely.
    """
    names = ALL_TOOL_NAMES if allowed is None else allowed
    declarations = []

    for tool in TOOLS:
        if tool.name not in names:
            continue

        raw_schema = tool.args_schema.model_json_schema()
        clean = _clean_schema(raw_schema)

        declarations.append(
            types.FunctionDeclaration(
                name=tool.name,
                description=tool.description,
                parameters=clean,
            )
        )

    if not declarations:
        return None

    return types.Tool(function_declarations=declarations)


async def dispatch_tool(
    name: str,
    args: dict,
    allowed: frozenset[str] | None = None,
) -> str:
    """
    Validate and execute one tool call, returning a string for the conversation.

    The allowlist is re-checked here rather than trusted from the declarations,
    because a model can call a function name it was never offered.
    """
    names = ALL_TOOL_NAMES if allowed is None else allowed

    tool = TOOL_MAP.get(name)
    if not tool or name not in names:
        return (
            f"Error: Tool '{name}' is not available to this agent. "
            f"Available tools: {sorted(names)}"
        )

    try:
        tool.args_schema(**args)
    except ValidationError as e:
        return f"Error: Invalid arguments for tool '{name}': {e}"

    try:
        result = await tool.execute(**args)
        return str(result)
    except Exception as e:
        return f"Error executing tool '{name}': {e}"
