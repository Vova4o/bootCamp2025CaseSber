from typing import Union
from .llm_client import LLMClient
from .openai_client import OpenAIClient
import logging

logger = logging.getLogger(__name__)

def create_llm_client(
    provider: str,
    api_key: str,
    base_url: str = None,
    model: str = None
) -> Union[LLMClient, OpenAIClient]:
    """
    Фабрика для создания LLM клиента
    
    Args:
        provider: "local" или "openai"
        api_key: API ключ
        base_url: URL для локальной модели (только для local)
        model: Название модели для OpenAI
    """
    provider = provider.lower()
    
    if provider == "openai":
        logger.info(f"Инициализация OpenAI клиента с моделью: {model}")
        return OpenAIClient(
            api_key=api_key,
            model=model or "gpt-4-turbo-preview"
        )
    elif provider == "local":
        logger.info(f"Инициализация локального LLM клиента: {base_url}")
        return LLMClient(
            base_url=base_url,
            api_key=api_key or "dummy"
        )
    else:
        raise ValueError(
            f"Неизвестный провайдер: {provider}. "
            f"Используйте 'local' или 'openai'"
        )