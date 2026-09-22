from google.genai import types
from pydantic import ValidationError

from tools.scrapper import ScrapperTool
from tools.search import WebSearchTool
from tools.emailer import EmailTool

# The full registry. A run never sees this directly — it gets a per-run subset
# resolved by resolve_allowed(), so an agent configured with only web_search
# cannot reach send_email.
TOOLS = [
    WebSearchTool(),
    ScrapperTool(),
    EmailTool(),
]

# name → instance lookup
TOOL_MAP = {t.name: t for t in TOOLS}

ALL_TOOL_NAMES = frozenset(TOOL_MAP)


#  Allowlist resolution

def resolve_allowed(
    requested: list[str] | None,
    email: str | None = None,
) -> frozenset[str]:
    """
    Turn an agent's configured tool list into the set it may actually call.

    - None  → every registered tool (back-compat for bare {"goal": "..."} calls)
    - []    → no tools; the model answers from its own knowledge
    - list  → intersected with the registry, so an unknown name in the database
              is dropped instead of crashing the run

    When `email` is set, send_email is added regardless. Delivery is the point
    of a scheduled briefing, and an agent whose allowlist omits it would
    research successfully and then silently fail to deliver.
    """
    if requested is None:
        allowed = set(ALL_TOOL_NAMES)
    else:
        allowed = {name for name in requested if name in TOOL_MAP}

    if email:
        allowed.add(EmailTool.name)

    return frozenset(allowed)


#  Schema conversion

def _clean_schema(schema: dict) -> dict:
    """
    Strip 'title' fields that Pydantic adds but Gemini doesn't accept.
    Gemini's FunctionDeclaration parameters must be a clean JSON Schema object.
    """
    schema.pop("title", None)
    for prop in schema.get("properties", {}).values():
        prop.pop("title", None)
    return schema


def get_gemini_tool_declarations(allowed: frozenset[str] | None = None) -> types.Tool | None:
    """
    Convert the allowed BaseTool instances into a single Gemini types.Tool.

    Returns None when nothing is allowed. Gemini rejects a Tool carrying an
    empty function_declarations list, so the caller must omit `tools` from the
    request config entirely in that case.
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
    Look up a tool by name, check it against the allowlist, validate its
    arguments, and execute it. Always returns a string (safe to append back to
    the Gemini conversation).

    The allowlist is re-checked here rather than trusted from the declarations:
    a model can hallucinate a function name it was never offered, and that call
    must not reach a real tool.
    """
    names = ALL_TOOL_NAMES if allowed is None else allowed

    tool = TOOL_MAP.get(name)
    if not tool or name not in names:
        return (
            f"Error: Tool '{name}' is not available to this agent. "
            f"Available tools: {sorted(names)}"
        )

    # Validate args against the tool's Pydantic schema
    try:
        tool.args_schema(**args)
    except ValidationError as e:
        return f"Error: Invalid arguments for tool '{name}': {e}"

    try:
        result = await tool.execute(**args)
        return str(result)
    except Exception as e:
        return f"Error executing tool '{name}': {e}"
