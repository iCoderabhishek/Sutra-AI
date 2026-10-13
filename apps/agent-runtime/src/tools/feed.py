"""read_feed: latest items from an RSS 2.0 or Atom feed, with real dates and links."""

import html
import re
import xml.etree.ElementTree as ET
from typing import Any, Type

import httpx
from pydantic import BaseModel, Field

from tools.base import BaseTool
from tools.net import UnsafeURL, fetch_text

ATOM = "{http://www.w3.org/2005/Atom}"
SUMMARY_CHARS = 300
_TAGS = re.compile(r"<[^>]+>")


class ReadFeedArgs(BaseModel):
    url: str = Field(description="Full URL of an RSS or Atom feed")
    limit: int = Field(default=10, ge=1, le=25, description="How many recent items to return")


def _text(el: ET.Element | None) -> str:
    """Plain text from an element: HTML tags removed, entities decoded, whitespace collapsed."""
    if el is None or el.text is None:
        return ""
    text = html.unescape(_TAGS.sub(" ", el.text))
    return " ".join(text.split())


def _short(text: str) -> str:
    return text if len(text) <= SUMMARY_CHARS else text[: SUMMARY_CHARS - 1].rstrip() + "…"


def _parse(xml: str) -> tuple[str, list[dict]]:
    root = ET.fromstring(xml)

    # RSS 2.0: <rss><channel><item>
    channel = root.find("channel")
    if channel is not None:
        items = [
            {
                "title": _text(i.find("title")),
                "link": _text(i.find("link")),
                "date": _text(i.find("pubDate")),
                "summary": _text(i.find("description")),
            }
            for i in channel.findall("item")
        ]
        return _text(channel.find("title")), items

    # Atom: <feed><entry>
    if root.tag == f"{ATOM}feed":
        items = []
        for e in root.findall(f"{ATOM}entry"):
            link = e.find(f"{ATOM}link[@rel='alternate']")
            if link is None:
                link = e.find(f"{ATOM}link")
            items.append({
                "title": _text(e.find(f"{ATOM}title")),
                "link": link.get("href", "") if link is not None else "",
                "date": _text(e.find(f"{ATOM}published")) or _text(e.find(f"{ATOM}updated")),
                "summary": _text(e.find(f"{ATOM}summary")) or _text(e.find(f"{ATOM}content")),
            })
        return _text(root.find(f"{ATOM}title")), items

    raise ValueError("not an RSS or Atom feed")


class ReadFeedTool(BaseTool):
    name: str = "read_feed"
    description: str = (
        "Read the latest items from an RSS or Atom feed URL (news sites, blogs, "
        "release notes, podcasts). Returns each item's title, date, link and short "
        "summary. Prefer this over web search when the source has a feed."
    )
    args_schema: Type[BaseModel] = ReadFeedArgs

    async def execute(self, **kwargs) -> Any:
        args = ReadFeedArgs(**kwargs)
        try:
            final_url, body = await fetch_text(args.url)
            title, items = _parse(body)
        except UnsafeURL as e:
            return f"Error: refused to fetch this URL: {e}"
        except httpx.HTTPError as e:
            return f"Error: could not fetch the feed: {e}"
        except (ET.ParseError, ValueError) as e:
            return f"Error: this URL is not a readable RSS or Atom feed ({e})"

        items = [i for i in items if i["title"] or i["link"]][: args.limit]
        if not items:
            return f"The feed at {final_url} has no items."

        lines = [f"Feed: {title or final_url} ({final_url})", f"Latest {len(items)} items:"]
        for n, i in enumerate(items, 1):
            lines.append(f"\n{n}. {i['title'] or '(untitled)'}")
            if i["date"]:
                lines.append(f"   Date: {i['date']}")
            if i["link"]:
                lines.append(f"   Link: {i['link']}")
            if i["summary"]:
                lines.append(f"   Summary: {_short(i['summary'])}")
        return "\n".join(lines)
