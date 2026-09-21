import redis.asyncio as aioredis

from libs.env import settings

redis_client = aioredis.from_url(settings.REDIS_URL, decode_responses=True)
