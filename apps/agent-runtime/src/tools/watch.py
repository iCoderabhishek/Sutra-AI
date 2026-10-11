"""
watch_page: compare a page with the snapshot this agent took last time and
return what changed (pricing pages, changelogs, job boards, docs).

Snapshots are per agent and per URL in Redis. Unlike memory, this is an exact
text diff, so it doesn't depend on what the model chose to write down.

Known limits:
- Pages that render their content with JavaScript come back mostly empty.
- Pages with timestamps, counters or rotating banners always look "changed".
"""

import difflib
import hashlib
import json
from datetime import datetime, timezone
from html.parser import HTMLParser
from typing import Any, Type

import httpx
from pydantic import BaseModel, Field

from libs.redis_client import redis_client
from tools.base import BaseTool
from tools.memory import current_agent_id
from tools.net import UnsafeURL, fetch_text

TTL_SECONDS = 90 * 24 * 60 * 60
MAX_SNAPSHOT_CHARS = 200_000
MAX_DIFF_LINES = 60
MAX_LINE_CHARS = 300

_SKIP = {"script", "style", "noscript", "svg", "template", "head"}
_BLOCK = {
    "p", "div", "section", "article", "li", "tr", "br", "h1", "h2", "h3", "h4",
    "h5", "h6", "header", "footer", "main", "td", "th", "table", "ul", "ol",
}


class _TextExtractor(HTMLParser):
    """HTML to plain text: one line per block element, scripts and styles dropped."""

    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.parts: list[str] = []
        self.skip_depth = 0

    def handle_starttag(self, tag, attrs):
        if tag in _SKIP:
            self.skip_depth += 1
        elif tag in _BLOCK:
            self.parts.append("\n")

    def handle_endtag(self, tag):
        if tag in _SKIP and self.skip_depth:
            self.skip_depth -= 1
        elif tag in _BLOCK:
            self.parts.append("\n")

    def handle_data(self, data):
        if not self.skip_depth:
            self.parts.append(data)


def page_text(html: str) -> str:
    parser = _TextExtractor()
    parser.feed(html)
    lines = (" ".join(line.split()) for line in "".join(parser.parts).split("\n"))
    return "\n".join(line for line in lines if line)[:MAX_SNAPSHOT_CHARS]


def _key(agent_id: str, url: str) -> str:
    digest = hashlib.sha256(url.encode()).hexdigest()[:24]
    return f"sutra:agent:{agent_id}:watch:{digest}"


def _clip(line: str) -> str:
    return line if len(line) <= MAX_LINE_CHARS else line[: MAX_LINE_CHARS - 1] + "…"


def describe_change(old: str, new: str) -> tuple[int, int, list[str]]:
    """Return (added, removed, diff lines) between two snapshots."""
    added, removed, lines = 0, 0, []
    for line in difflib.unified_diff(old.split("\n"), new.split("\n"), lineterm="", n=0):
        if line.startswith(("+++", "---", "@@")):
            continue
        if line.startswith("+"):
            added += 1
            lines.append("+ " + _clip(line[1:]))
        elif line.startswith("-"):
            removed += 1
            lines.append("- " + _clip(line[1:]))
    return added, removed, lines


class WatchPageArgs(BaseModel):
    url: str = Field(description="Full URL of the page to watch")


class WatchPageTool(BaseTool):
    name: str = "watch_page"
    description: str = (
        "Check a web page for changes since this agent last looked at it. Returns "
        "the exact lines added and removed (e.g. a price or a new changelog entry), "
        "or says nothing changed. The first call saves a baseline snapshot."
    )
    args_schema: Type[BaseModel] = WatchPageArgs

    async def execute(self, **kwargs) -> Any:
        args = WatchPageArgs(**kwargs)
        agent_id = current_agent_id.get()
        if not agent_id:
            return "Error: page watching is not available for this run."

        try:
            final_url, body = await fetch_text(args.url)
        except UnsafeURL as e:
            return f"Error: refused to fetch this URL: {e}"
        except httpx.HTTPError as e:
            return f"Error: could not fetch the page: {e}"

        text = page_text(body)
        if not text:
            return f"The page at {final_url} has no readable text (it may need JavaScript)."

        now = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
        digest = hashlib.sha256(text.encode()).hexdigest()
        key = _key(agent_id, args.url)

        raw = await redis_client.get(key)
        previous = json.loads(raw) if raw else None
        await redis_client.set(key, json.dumps({"at": now, "hash": digest, "text": text}), ex=TTL_SECONDS)

        if previous is None:
            return (
                f"First check of {final_url}: baseline saved ({len(text.splitlines())} lines). "
                "There is nothing to compare yet; changes will show from the next run."
            )

        if previous.get("hash") == digest:
            return f"No change on {final_url} since {previous.get('at')}."

        added, removed, lines = describe_change(previous.get("text", ""), text)
        shown = lines[:MAX_DIFF_LINES]
        more = f"\n… {len(lines) - len(shown)} more changed lines not shown" if len(lines) > len(shown) else ""
        return (
            f"{final_url} changed since {previous.get('at')}: "
            f"{added} lines added, {removed} removed.\n"
            "Lines starting with + are new, - were removed:\n"
            + "\n".join(shown) + more
        )
