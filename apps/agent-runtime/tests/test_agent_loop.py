import asyncio
import os
import sys

# Add the parent directory's src folder to the python path
sys.path.append(os.path.join(os.path.dirname(os.path.dirname(__file__)), "src"))

from agent.loop import run_agent

async def mock_stream(event: dict):
    print(f"[{event.get('status', 'INFO').upper()}] {event.get('step')} - Cost so far: ${event.get('cost', {}).get('cost_usd', 0):.4f}")
    if 'result_preview' in event:
        print(f"  -> Preview: {event['result_preview'].strip()}")
    if 'args' in event:
        print(f"  -> Args: {event['args']}")
    if 'content' in event and event['step'] == 'Final Answer':
        print(f"\n--- FINAL ANSWER ---\n{event['content']}")

async def main():
    print("Starting Agent Run Test...")
    goal = "Search the web for the latest news on Anthropic and generate a very brief summary. Then use the scrapper tool to read the main anthropic.com page to see their current tagline, and finally write a short report."
    
    # We will give it web_search and scrapper. We won't use email to avoid spamming the user's outbox if they really put credentials, unless they want us to.
    tools = ["web_search", "scrapper"]
    
    try:
        await run_agent(
            goal=goal,
            stream_callback=mock_stream,
            tools=tools,
            system_prompt="You are a test agent made by Abhishek OG developer. Please execute the goal precisely and in a cool and concise manner.",
            template=None,
            instruction="Keep it brief and concise.",
            email=None
        )
    except Exception as e:
        print(f"Agent loop failed: {e}")

if __name__ == "__main__":
    asyncio.run(main())
