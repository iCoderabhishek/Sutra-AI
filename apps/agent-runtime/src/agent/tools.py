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


def get_gemini_tool_declarations() -> types.Tool:
    """
    Converts all registered BaseTool instances into a single Gemini types.Tool
    containing one FunctionDeclaration per tool.
    """
    declarations = []

    for tool in TOOLS:
        raw_schema = tool.args_schema.model_json_schema()
        clean = _clean_schema(raw_schema)

        declaration = types.FunctionDeclaration(
            name=tool.name,
            description=tool.description,
            parameters=clean,
        )
        declarations.append(declaration)

    return types.Tool(function_declarations=declarations)


async def dispatch_tool(name: str, args: dict) -> str:
    """
    Look up a tool by name, validate its arguments, and execute it.
    Always returns a string (safe to append back to the Gemini conversation).
    """
    tool = TOOL_MAP.get(name)
    if not tool:
        return f"Error: Unknown tool '{name}'. Available tools: {list(TOOL_MAP.keys())}"

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
