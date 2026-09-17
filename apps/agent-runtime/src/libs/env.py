from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    # API Keys
    TAVILY_API_KEY: str = ""
    OPENAI_API_KEY: str = ""
    GEMINI_API_KEY: str = ""
    FIRECRAWL_API_KEY: str = ""
    RESEND_API_KEY: str = ""

   
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

settings = Settings()
