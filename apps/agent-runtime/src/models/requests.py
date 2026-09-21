from pydantic import BaseModel
from typing import Optional


class RunRequest(BaseModel):
    goal: str
    tools: Optional[list[str]] = None
    email: Optional[str] = None
    run_id: Optional[str] = None
