import json
import urllib.request

def get_free_models():
    req = urllib.request.Request("https://openrouter.ai/api/v1/models")
    with urllib.request.urlopen(req) as response:
        data = json.loads(response.read().decode())
        free_models = []
        for model in data["data"]:
            pricing = model.get("pricing", {})
            prompt = pricing.get("prompt")
            completion = pricing.get("completion")
            if prompt == "0" and completion == "0":
                free_models.append(model["id"])
        
        print("Available Free Models:")
        for m in sorted(free_models):
            print(m)

if __name__ == "__main__":
    get_free_models()
