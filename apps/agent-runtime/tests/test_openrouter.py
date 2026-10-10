import asyncio
from libs.openrouter_client import generate_content_openai, FREE_MODELS

async def test_openrouter():
    print("Testing OpenRouter connection...")
    
    messages = [
        {"role": "user", "content": "What is the capital of France? Reply in exactly one word."}
    ]
    
    for model in FREE_MODELS:
        kwargs = {
            "model": model,
            "messages": messages,
            "temperature": 0.0,
        }
        try:
            print(f"Testing model: {model}")
            response = await generate_content_openai(**kwargs)
            content = response.choices[0].message.content
            print(f"[SUCCESS] {model} returned: {content}")
        except Exception as e:
            print(f"[ERROR] {model} failed with: {e}")

if __name__ == "__main__":
    asyncio.run(test_openrouter())
