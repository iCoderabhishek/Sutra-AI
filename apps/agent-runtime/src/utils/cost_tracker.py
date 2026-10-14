
from utils.token_counter import extract_usage

# AWS Bedrock pricing per 1M tokens
PRICING = {
    "anthropic.claude-3-5-sonnet-20240620-v1:0": {"input": 3.00,  "output": 15.00},
    "us.anthropic.claude-3-5-sonnet-20241022-v2:0": {"input": 3.00,  "output": 15.00},
}


class CostTracker:
    """
    Tracks token usage and cost for a single agent run.
    Call add() after each Gemini response, then total() to get the summary.
    """

    def __init__(self, model: str = "meta-llama/llama-3.3-70b-instruct:free"):
        self.model = model
        self.input_tokens = 0
        self.output_tokens = 0

    def add(self, response) -> None:
        usage = extract_usage(response)
        self.input_tokens  += usage["input_tokens"]
        self.output_tokens += usage["output_tokens"]

    def total(self) -> dict:
        pricing = PRICING.get(self.model, {"input": 0, "output": 0})
        cost = (self.input_tokens / 1_000_000) * pricing["input"] \
             + (self.output_tokens / 1_000_000) * pricing["output"]

        return {
            "input_tokens":  self.input_tokens,
            "output_tokens": self.output_tokens,
            "total_tokens":  self.input_tokens + self.output_tokens,
            "cost_usd":      round(cost, 6),
            "model":         self.model,
        }
