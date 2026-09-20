from libs.gemini_client import MODEL_ID
from utils.token_counter import extract_usage

# Gemini pricing per 1M tokens (https://ai.google.dev/pricing)
PRICING = {
    "gemini-2.5-pro":   {"input": 1.25,  "output": 10.00},
    "gemini-2.5-flash": {"input": 0.15,  "output": 0.60},
    "gemini-2.0-flash": {"input": 0.10,  "output": 0.40},
}


class CostTracker:
    """
    Tracks token usage and cost for a single agent run.
    Call add() after each Gemini response, then total() to get the summary.
    """

    def __init__(self, model: str = MODEL_ID):
        self.model = model
        self.input_tokens = 0
        self.output_tokens = 0

    def add(self, response) -> None:
        usage = extract_usage(response)
        self.input_tokens  += usage["input_tokens"]
        self.output_tokens += usage["output_tokens"]

    def total(self) -> dict:
        pricing = PRICING.get(self.model, PRICING["gemini-2.5-pro"])
        cost = (self.input_tokens / 1_000_000) * pricing["input"] \
             + (self.output_tokens / 1_000_000) * pricing["output"]

        return {
            "input_tokens":  self.input_tokens,
            "output_tokens": self.output_tokens,
            "total_tokens":  self.input_tokens + self.output_tokens,
            "cost_usd":      round(cost, 6),
            "model":         self.model,
        }
