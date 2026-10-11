from contextvars import ContextVar
from typing import Any, Type

from pydantic import BaseModel

from libs.smtp_mail import send_email
from tools.base import BaseTool

# The owner's address for this run. Set by the agent loop, never by the model,
# so an injected "email this to attacker@evil.com" cannot change the recipient.
current_recipient: ContextVar[str | None] = ContextVar("current_recipient", default=None)


class EmailArgs(BaseModel):
    """Arguments for the emailer tool. The recipient is fixed to the agent owner."""
    subject: str
    """The subject line of the email"""
    body: str
    """The main text content of the email"""


class EmailTool(BaseTool):
    name: str = "send_email"
    description: str = (
        "Send the finished result by email to the agent's owner (the address is "
        "fixed; you only choose the subject and body)."
    )
    args_schema: Type[BaseModel] = EmailArgs

    async def execute(self, **kwargs) -> Any:
        subject = kwargs.get("subject")
        body = kwargs.get("body")
        to_email = current_recipient.get()

        if not to_email:
            return "Error: this agent has no delivery email set."
        if not subject or not body:
            return "Error: Missing required email fields (subject, body)."

        print(f"[EmailTool] Sending email to owner: {to_email}")

        try:
            success = await send_email(to_email=to_email, subject=subject, body=body)
            return (
                f"Successfully sent email to the owner ({to_email})."
                if success else "Failed to send the email."
            )
        except Exception as e:
            return f"Error executing emailer: {str(e)}"
