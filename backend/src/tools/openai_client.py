import httpx
from typing import List, Dict
import logging

logger = logging.getLogger(__name__)

class OpenAIClient:
    def __init__(self, api_key: str, model: str = "gpt-4-turbo-preview"):
        self.api_key = api_key
        self.model = model
        self.base_url = "https://api.openai.com/v1"
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
                    "model": self.model,
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
            logger.error(f"OpenAI API Error: {e}")
            if hasattr(e, 'response'):
                logger.error(f"Response: {e.response.text}")
            return f"Ошибка при обращении к OpenAI: {str(e)}"
        except Exception as e:
            logger.error(f"Unexpected error: {e}")
            return f"Неожиданная ошибка: {str(e)}"
    
    async def close(self):
        await self.client.aclose()