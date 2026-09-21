from pydantic import BaseModel, EmailStr
from typing import Optional


class RunRequest(BaseModel):
    """Payload for POST /run — what the Node.js backend sends to kick off an agent."""
    goal: str
    email: Optional[str] = None 
    run_id: Optional[str] = None 
