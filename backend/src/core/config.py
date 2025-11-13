from pydantic_settings import BaseSettings
from functools import lru_cache
from typing import List
import json

class Settings(BaseSettings):
    # Database
    database_url: str = "postgresql://research_user:research_pass@localhost:5432/research_db"
    
    # Redis
    redis_url: str = "redis://localhost:6379"
    
    # Search Engine
    search_engine: str = "duckduckgo"  # duckduckgo, tavily, brave
    search_region: str = "wt-wt"  # ru-ru для России, wt-wt для мира
    
    # LLM Provider
    llm_provider: str = "local"
    
    # Local LLM
    llm_api_url: str = "http://your-clore-server-ip:5000/v1"
    llm_api_key: str = "dummy"
    
    # OpenAI
    openai_api_key: str = ""
    openai_model: str = "gpt-4-turbo-preview"
    
    # Router settings
    use_llm_router: bool = True
    
    # Search settings
    max_results_simple: int = 5
    max_results_pro: int = 10
    max_content_length: int = 2500
    
    # Context settings
    max_context_messages: int = 10
    max_context_tokens: int = 4000
    
    # CORS
    cors_origins: List[str] = ["http://localhost:3000"]
    
    # Server
    host: str = "0.0.0.0"
    port: int = 8080
    
    @property
    def get_cors_origins(self) -> List[str]:
        """Преобразует строку в список если нужно"""
        if isinstance(self.cors_origins, str):
            if self.cors_origins.startswith('['):
                try:
                    return json.loads(self.cors_origins)
                except:
                    pass
            return [origin.strip() for origin in self.cors_origins.split(',')]
        return self.cors_origins
    
    class Config:
        env_file = ".env"
        case_sensitive = False

@lru_cache()
def get_settings() -> Settings:
    return Settings()

settings = get_settings()