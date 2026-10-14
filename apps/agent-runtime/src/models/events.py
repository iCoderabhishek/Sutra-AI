from pydantic import BaseModel
from typing import Any, Optional
from enum import Enum


class TraceStatus(str, Enum):
    running = "running"
    done    = "done"
    error   = "error"


class TraceEvent(BaseModel):
    """
    A single step emitted by the agent loop via stream_callback.
    Every SSE message is one of these, JSON-serialized.
    """
    step:           str
    status:         TraceStatus
    iteration:      Optional[int]  = None
    content:        Optional[str]  = None  
    args:           Optional[dict] = None  
    result_preview: Optional[str]  = None  # first 200 chars of tool result
    cost:           Optional[dict] = None  # CostTracker.total() snapshot
