import httpx
from typing import List, Dict
import logging

logger = logging.getLogger(__name__)

class LLMClient:
    def __init__(self, base_url: str, api_key: str = "dummy"):
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.client = httpx.AsyncClient(timeout=60.0)
    
    async def chat_completion(
        self,
        messages: List[Dict[str, str]],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        stream: bool = False
    ) -> str:
        try:
            response = await self.client.post(
                f"{self.base_url}/chat/completions",
                json={
                    "model": "local-model",
                    "messages": messages,
                    "temperature": temperature,
                    "max_tokens": max_tokens,
                    "stream": stream
                },
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json"
                }
            )
            response.raise_for_status()
            
            result = response.json()
            return result["choices"][0]["message"]["content"]
            
        except httpx.HTTPError as e:
            logger.error(f"LLM API Error: {e}")
            return f"Ошибка при обращении к LLM: {str(e)}"
        except Exception as e:
            logger.error(f"Unexpected error: {e}")
            return f"Неожиданная ошибка: {str(e)}"
    
    async def close(self):
        await self.client.aclose()