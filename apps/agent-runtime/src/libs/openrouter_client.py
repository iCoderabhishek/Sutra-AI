import urllib.request
import json
from openai import AsyncOpenAI
from libs.env import settings

openrouter_client = AsyncOpenAI(
    base_url="https://openrouter.ai/api/v1",
    api_key=settings.OPENROUTER_API_KEY,
)

_free_models_cache = None

def get_free_models():
    global _free_models_cache
    if _free_models_cache is not None:
        return _free_models_cache
        
    try:
        req = urllib.request.Request("https://openrouter.ai/api/v1/models")
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
            free_models = []
            for model in data.get("data", []):
                pricing = model.get("pricing", {})
                prompt = pricing.get("prompt")
                completion = pricing.get("completion")
                if prompt == "0" and completion == "0":
                    free_models.append(model["id"])
            

            if "openrouter/free" in free_models:
                free_models.remove("openrouter/free")
            free_models.insert(0, "openrouter/free")
            
            _free_models_cache = free_models
            return free_models
    except Exception as e:
        print(f"Failed to fetch free models: {e}")
        return ["openrouter/free"]

AGENT_TEMPERATURE = 0.3

async def generate_content_openai(**kwargs):
    return await openrouter_client.chat.completions.create(**kwargs)
