from libs.env import settings
from tavily import AsyncTavilyClient

# Initialize a single reusable client
tavily_client = AsyncTavilyClient(api_key=settings.TAVILY_API_KEY)

# I am using async client to maintain the non blocking nature of async fast api