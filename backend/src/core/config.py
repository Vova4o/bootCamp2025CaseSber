from pydantic_settings import BaseSettings
from functools import lru_cache
from typing import List

class Settings(BaseSettings):
    # Database
    database_url: str = "postgresql://research_user:research_pass@localhost:5432/research_db"
    
    # Redis
    redis_url: str = "redis://localhost:6379"
    
    # Search
    tavily_url: str = "http://localhost:8000"
    searxng_url: str = "http://localhost:8999"
    
    # LLM Provider
    llm_provider: str = "local"
    
    # Local LLM
    llm_api_url: str = "http://your-clore-server-ip:5000/v1"
    llm_api_key: str = "dummy"
    
    # OpenAI
    openai_api_key: str = ""
    openai_model: str = "gpt-4-turbo-preview"
    
    # Router settings
    use_llm_router: bool = True  # Использовать LLM для классификации
    
    # Search settings
    max_results_simple: int = 5
    max_results_pro: int = 10
    scraping_timeout: int = 10
    max_content_length: int = 2500
    
    # Context settings
    max_context_messages: int = 10
    max_context_tokens: int = 4000
    
    # ML Models
    sentence_model: str = "all-MiniLM-L6-v2"
    
    # CORS
    cors_origins: List[str] = ["http://localhost:3000"]
    
    # Server
    host: str = "0.0.0.0"
    port: int = 8080
    
    class Config:
        env_file = ".env"
        case_sensitive = False

@lru_cache()
def get_settings() -> Settings:
    return Settings()

settings = get_settings()