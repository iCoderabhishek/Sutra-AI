def extract_usage(response) -> dict:
    """
    Reads token counts from Gemini's response.usage_metadata.
    No tiktoken needed — Gemini reports its own token counts.

    Returns:
        {
            "input_tokens":  int,
            "output_tokens": int,
            "total_tokens":  int,
        }
    """
    meta = getattr(response, "usage_metadata", None)
    if meta is None:
        return {"input_tokens": 0, "output_tokens": 0, "total_tokens": 0}

    input_tokens  = getattr(meta, "prompt_token_count",     0) or 0
    output_tokens = getattr(meta, "candidates_token_count", 0) or 0
    total_tokens  = getattr(meta, "total_token_count",      0) or 0

    return {
        "input_tokens":  input_tokens,
        "output_tokens": output_tokens,
        "total_tokens":  total_tokens,
    }
