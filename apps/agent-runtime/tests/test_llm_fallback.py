import asyncio
import os
import sys

# Add src directory to PYTHONPATH so we can import from our project
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../src')))

from agent.loop import _call_with_fallback
from libs.openrouter_client import get_free_models

async def test_fallback():
    print("="*60)
    print("LLM FALLBACK MECHANISM TEST")
    print("="*60)
    
    models = get_free_models()
    print(f"Dynamically fetched free models: {models[:5]}... (showing first 5)")
    
    kwargs = {
        "messages": [{"role": "user", "content": "What is 2+2? Reply with just the number."}],
        "temperature": 0.0
    }
    
    # 1. Test normal execution
    print("\n[Test 1] Normal execution (First model should respond)")
    try:
        response = await _call_with_fallback(kwargs.copy())
        print(f"✅ SUCCESS! Model used: {response.model}")
        print(f"   Response: {response.choices[0].message.content}")
    except Exception as e:
        print(f"❌ FAILED: {e}")

    # 2. Test fallback by temporarily breaking the first few models
    print("\n[Test 2] Fallback execution (Simulating API failures)")
    
    import libs.openrouter_client as or_client
    
    # Force the cache to use bad models first
    original_models = or_client._free_models_cache.copy()
    
    try:
        or_client._free_models_cache = [
            "fake-provider/model-that-does-not-exist:free",
            "fake-provider/another-bad-model:free"
        ] + original_models
        
        print(f"   Injected fake models. Current fallback chain: {or_client._free_models_cache[:4]}...")
        print("   Calling API... (You should see logs showing it trying and switching)")
        
        response = await _call_with_fallback(kwargs.copy())
        
        print(f"✅ FALLBACK SUCCESS! The system successfully switched to: {response.model}")
        print(f"   Response: {response.choices[0].message.content}")
        
    except Exception as e:
        print(f"❌ FALLBACK FAILED: {e}")
    finally:
        # Restore the original list
        or_client._free_models_cache = original_models

if __name__ == "__main__":
    asyncio.run(test_fallback())
