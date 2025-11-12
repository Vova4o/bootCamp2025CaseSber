import httpx
from typing import Dict
import logging

logger = logging.getLogger(__name__)

class SearchClient:
    def __init__(self, tavily_url: str):
        self.tavily_url = tavily_url.rstrip('/')
        self.client = httpx.AsyncClient(timeout=30.0)
    
    async def search(
        self,
        query: str,
        max_results: int = 5,
        include_raw_content: bool = False
    ) -> Dict:
        try:
            response = await self.client.post(
                f"{self.tavily_url}/search",
                json={
                    "query": query,
                    "max_results": max_results,
                    "include_raw_content": include_raw_content
                }
            )
            response.raise_for_status()
            return response.json()
            
        except httpx.HTTPError as e:
            logger.error(f"Search API Error: {e}")
            return {"results": [], "error": str(e)}
        except Exception as e:
            logger.error(f"Unexpected error: {e}")
            return {"results": [], "error": str(e)}
    
    async def close(self):
        await self.client.aclose()