"""
memory: notes an agent keeps between runs, so a scheduled agent can report only
what changed since last time.

Storage is a capped Redis list per agent (newest first). The agent ID comes
from the run itself (current_agent_id), never from the model, so one agent
can't read or write another agent's memory.
"""

import json
from contextvars import ContextVar
from datetime import datetime, timezone
from typing import Any, Literal, Type

from pydantic import BaseModel, Field

from agent.guardrails import clean_text, redact
from libs.redis_client import redis_client
from tools.base import BaseTool

MAX_NOTES = 50
MAX_NOTE_CHARS = 800
RECALL_DEFAULT = 5
TTL_SECONDS = 90 * 24 * 60 * 60

# Set by the agent loop for the run in progress.
current_agent_id: ContextVar[str | None] = ContextVar("current_agent_id", default=None)


def _key(agent_id: str) -> str:
    return f"sutra:agent:{agent_id}:memory"


async def save_note(agent_id: str, note: str) -> str:
    note = redact(clean_text(note)).strip()[:MAX_NOTE_CHARS]
    entry = json.dumps({"at": datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC"), "note": note})
    key = _key(agent_id)
    pipe = redis_client.pipeline()
    pipe.lpush(key, entry)
    pipe.ltrim(key, 0, MAX_NOTES - 1)
    pipe.expire(key, TTL_SECONDS)
    await pipe.execute()
    return note


async def recent_notes(agent_id: str, limit: int = RECALL_DEFAULT) -> list[dict]:
    raw = await redis_client.lrange(_key(agent_id), 0, limit - 1)
    notes = []
    for item in raw:
        try:
            notes.append(json.loads(item))
        except json.JSONDecodeError:
            continue
    return notes


def format_notes(notes: list[dict]) -> str:
    if not notes:
        return "No memory yet: this is the agent's first run."
    lines = [f"- [{n.get('at', '?')}] {n.get('note', '')}" for n in notes]
    return "Notes from previous runs, newest first:\n" + "\n".join(lines)


class MemoryArgs(BaseModel):
    action: Literal["save", "recall"] = Field(description="'save' a note for future runs, or 'recall' recent notes")
    note: str | None = Field(default=None, description="For 'save': the key facts to remember, with dates and URLs")
    limit: int = Field(default=RECALL_DEFAULT, ge=1, le=20, description="For 'recall': how many notes")


class MemoryTool(BaseTool):
    name: str = "memory"
    description: str = (
        "Notes this agent keeps between its runs. 'recall' returns what earlier runs "
        "found; 'save' stores key facts for the next run. Use it to report only what "
        "changed since last time."
    )
    args_schema: Type[BaseModel] = MemoryArgs

    async def execute(self, **kwargs) -> Any:
        args = MemoryArgs(**kwargs)
        agent_id = current_agent_id.get()
        if not agent_id:
            return "Error: memory is not available for this run."

        if args.action == "recall":
            return format_notes(await recent_notes(agent_id, args.limit))

        if not args.note or not args.note.strip():
            return "Error: 'save' needs a non-empty note."
        saved = await save_note(agent_id, args.note)
        return f"Saved to memory ({len(saved)} characters)."
