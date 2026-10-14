def extract_usage(response) -> dict:
    """
    Reads token counts from OpenAI ChatCompletion API response.
    
    Returns:
        {
            "input_tokens":  int,
            "output_tokens": int,
            "total_tokens":  int,
        }
    """
    if hasattr(response, 'usage') and response.usage:
        return {
            "input_tokens":  response.usage.prompt_tokens or 0,
            "output_tokens": response.usage.completion_tokens or 0,
            "total_tokens":  response.usage.total_tokens or 0,
        }
    return {
        "input_tokens":  0,
        "output_tokens": 0,
        "total_tokens":  0,
    }
