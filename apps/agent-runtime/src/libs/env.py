from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    # API Keys
    TAVILY_API_KEY: str = ""
    OPENAI_API_KEY: str = ""
    GEMINI_API_KEY: str = ""
    FIRECRAWL_API_KEY: str = ""
    # Shared secret the Node backend must present on every non-health request.
    AGENT_SHARED_SECRET: str = ""
    # Redis
    REDIS_URL: str = "redis://localhost:6381"
    # SMTP Settings
    SMTP_HOST: str = ""
    SMTP_PORT: int = 587
    SMTP_USER: str = ""
    SMTP_PASSWORD: str = ""

    # Safety limits
    MAX_RUN_COST_USD: float = 0.25
    MAX_CONCURRENT_RUNS: int = 5

    # system prompts for ai
    SYSTEM_PROMPT: str = ""
    USER_PROMPT: str = ""
   
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

settings = Settings()
