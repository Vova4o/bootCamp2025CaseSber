from duckduckgo_search import DDGS
from typing import Dict, List
import logging
import asyncio
from concurrent.futures import ThreadPoolExecutor

logger = logging.getLogger(__name__)

class DuckDuckGoClient:
    """
    Асинхронный клиент для DuckDuckGo поиска
    """
    
    def __init__(self, max_workers: int = 3):
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
    
    async def search(
        self,
        query: str,
        max_results: int = 5,
        region: str = "wt-wt",  # wt-wt = worldwide
        safesearch: str = "moderate",  # off, moderate, strict
        include_raw_content: bool = False
    ) -> Dict:
        """
        Асинхронный поиск через DuckDuckGo
        
        Args:
            query: Поисковый запрос
            max_results: Количество результатов
            region: Регион поиска (ru-ru для России)
            safesearch: Безопасный поиск
            include_raw_content: Получить полный контент страниц
        
        Returns:
            {"results": [...], "query": "..."}
        """
        try:
            # DuckDuckGo библиотека синхронная, запускаем в executor
            loop = asyncio.get_event_loop()
            results = await loop.run_in_executor(
                self.executor,
                self._sync_search,
                query,
                max_results,
                region,
                safesearch
            )
            
            logger.info(f"DuckDuckGo found {len(results)} results for: {query}")
            
            return {
                "results": results,
                "query": query
            }
            
        except Exception as e:
            logger.error(f"DuckDuckGo search error: {e}")
            return {
                "results": [],
                "query": query,
                "error": str(e)
            }
    
    def _sync_search(
        self, 
        query: str, 
        max_results: int,
        region: str,
        safesearch: str
    ) -> List[Dict]:
        """Синхронный поиск (вызывается через executor)"""
        results = []
        
        try:
            with DDGS() as ddgs:
                search_results = ddgs.text(
                    keywords=query,
                    region=region,
                    safesearch=safesearch,
                    max_results=max_results
                )
                
                for item in search_results:
                    result = {
                        "title": item.get("title", ""),
                        "url": item.get("href", ""),
                        "content": item.get("body", ""),
                        "score": 1.0,
                    }
                    
                    # DuckDuckGo уже возвращает краткое содержимое
                    # Если нужен полный контент, можно добавить scraping
                    result["raw_content"] = result["content"]
                    
                    results.append(result)
        
        except Exception as e:
            logger.error(f"DuckDuckGo sync search error: {e}")
        
        return results
    
    async def close(self):
        """Закрытие executor"""
        self.executor.shutdown(wait=True)
        logger.info("DuckDuckGo client closed")