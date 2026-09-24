from pydantic import BaseModel
from typing import Optional


class RunRequest(BaseModel):
    """What the Node backend sends to start a run."""

    goal: str
    tools: Optional[list[str]] = None
    system_prompt: Optional[str] = None
    template: Optional[str] = None
    instruction: Optional[str] = None
    email: Optional[str] = None
    run_id: Optional[str] = None
