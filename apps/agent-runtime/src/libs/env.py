from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    # API Keys
    TAVILY_API_KEY: str = ""
    OPENAI_API_KEY: str = ""
    GEMINI_API_KEY: str = ""
    FIRECRAWL_API_KEY: str = ""
    # SMTP Settings
    SMTP_HOST: str = ""
    SMTP_PORT: int = 587
    SMTP_USER: str = ""
    SMTP_PASSWORD: str = ""

    # system prompts for ai
    SYSTEM_PROMPT: str = ""
    USER_PROMPT: str = ""
   
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

settings = Settings()
