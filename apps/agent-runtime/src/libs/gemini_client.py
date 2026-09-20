from google import genai
from google.genai import types
from libs.env import settings

gemini_client = genai.Client(api_key=settings.GEMINI_API_KEY)
MODEL_ID = 'gemini-2.5-pro'
TEMPERATURE = 0.7  
AGENT_TEMPERATURE = 0.3  

async def generate_content(prompt: str):
    response = await gemini_client.aio.models.generate_content(
        model=MODEL_ID,
        contents=prompt,
        config=types.GenerateContentConfig(
            temperature=TEMPERATURE,
            top_p=0.9,
        ),
    )
    return response.text