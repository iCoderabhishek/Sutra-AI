from pydantic import BaseModel
from typing import Optional


class RunRequest(BaseModel):
    """
    What the Node backend sends to start a run.

    Everything except `goal` is optional so a bare {"goal": "..."} still works,
    which keeps the health-probe and manual-curl paths simple.
    """

    goal: str
    tools: Optional[list[str]] = None
    system_prompt: Optional[str] = None
    template: Optional[str] = None
    instruction: Optional[str] = None
    email: Optional[str] = None
    run_id: Optional[str] = None
