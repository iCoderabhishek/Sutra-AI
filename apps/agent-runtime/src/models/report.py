"""
The fixed shape of every final answer.

The model is told to reply with exactly this JSON (see agent/prompts.py), the
loop validates it, and clients render it. Because the shape never changes,
every run looks the same in the CLI, the dashboard and stored traces.
"""

from typing import Literal, Optional

from pydantic import BaseModel, Field


class Finding(BaseModel):
    title: str
    points: list[str] = Field(default_factory=list)


class Source(BaseModel):
    name: str
    date: Optional[str] = None
    url: Optional[str] = None


class Report(BaseModel):
    outcome: Literal["completed", "partial", "declined"] = "completed"
    headline: str
    summary: str
    findings: list[Finding] = Field(default_factory=list)
    sources: list[Source] = Field(default_factory=list)
    takeaway: str = ""
    confidence: Literal["high", "medium", "low"] = "medium"
