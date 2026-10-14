"""
notify: send a short Telegram alert to the agent's owner.

The bot token is the runtime's (TELEGRAM_BOT_TOKEN), the chat ID comes from the
agent's config via the run (current_telegram_chat). The model only writes the
message, so it can't message anyone else.
"""

from contextvars import ContextVar
from typing import Any, Type

import httpx
from pydantic import BaseModel, Field

from agent.guardrails import clean_text, redact, strip_markdown
from libs.env import settings
from tools.base import BaseTool

MAX_MESSAGE_CHARS = 1000      # Telegram allows 4096; alerts should be short
MAX_SENDS_PER_RUN = 2
TIMEOUT_SECONDS = 10

# Set by the agent loop for the run in progress.
current_telegram_chat: ContextVar[str | None] = ContextVar("current_telegram_chat", default=None)
_sends: ContextVar[int] = ContextVar("notify_sends", default=0)


def reset_send_count() -> None:
    _sends.set(0)


class NotifyArgs(BaseModel):
    message: str = Field(description="Short plain-text alert: what changed and one source link. No markdown.")


class NotifyTool(BaseTool):
    name: str = "notify"
    description: str = (
        "Send a short Telegram alert to the agent's owner. Use it only when there is "
        "something new or important to tell them; skip it when nothing changed."
    )
    args_schema: Type[BaseModel] = NotifyArgs

    async def execute(self, **kwargs) -> Any:
        args = NotifyArgs(**kwargs)
        chat_id = current_telegram_chat.get()

        if not settings.TELEGRAM_BOT_TOKEN:
            return "Error: Telegram is not configured on this server."
        if not chat_id:
            return "Error: this agent has no Telegram chat ID set."
        if _sends.get() >= MAX_SENDS_PER_RUN:
            return f"Error: already sent {MAX_SENDS_PER_RUN} alerts this run. Put remaining details in the final report."

        text = redact(strip_markdown(clean_text(args.message)))[:MAX_MESSAGE_CHARS].strip()
        if not text:
            return "Error: the message is empty."

        url = f"https://api.telegram.org/bot{settings.TELEGRAM_BOT_TOKEN}/sendMessage"
        try:
            async with httpx.AsyncClient(timeout=TIMEOUT_SECONDS) as client:
                res = await client.post(url, json={
                    "chat_id": chat_id,
                    "text": text,
                    "disable_web_page_preview": True,
                })
            data = res.json()
        except (httpx.HTTPError, ValueError) as e:
            # httpx errors include the URL, which contains the bot token.
            return f"Error: could not reach Telegram ({redact(str(e))})."

        if not data.get("ok"):
            return f"Error: Telegram rejected the message: {redact(str(data.get('description', 'unknown error')))}"

        _sends.set(_sends.get() + 1)
        return "Alert sent on Telegram."
