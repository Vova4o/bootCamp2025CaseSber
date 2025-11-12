from pydantic_settings import BaseSettings
from functools import lru_cache
from typing import List

class Settings(BaseSettings):
    # Database
    database_url: str = "postgresql://research_user:research_pass@postgres:5432/research_db"
    
    # Redis
    redis_url: str = "redis://redis:6379"
    
    # Search
    tavily_url: str = "http://tavily-adapter:8000"
    searxng_url: str = "http://searxng:8080"
    
    # LLM
    llm_api_url: str
    llm_api_key: str = "dummy"
    
    # Search settings
    max_results_simple: int = 5
    max_results_pro: int = 10
    scraping_timeout: int = 10
    max_content_length: int = 2500
    
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