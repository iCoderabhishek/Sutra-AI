from firecrawl import AsyncFirecrawl
from libs.env import settings

firecrawl_client = AsyncFirecrawl(api_key=settings.FIRECRAWL_API_KEY)