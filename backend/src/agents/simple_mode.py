from typing import Dict
import time
import logging

logger = logging.getLogger(__name__)

async def process_simple_mode(
    query: str,
    search_client,
    llm_client,
    max_results: int = 5
) -> Dict:
    start_time = time.time()
    
    try:
        # 1. Поиск
        search_results = await search_client.search(
            query=query,
            max_results=max_results,
            include_raw_content=False
        )
        
        if not search_results.get("results"):
            return {
                "mode": "simple",
                "query": query,
                "answer": "Извините, не удалось найти информацию по вашему запросу.",
                "sources": [],
                "response_time": time.time() - start_time
            }
        
        # 2. Формирование контекста
        context = "\n\n".join([
            f"[{i+1}] {r.get('title', 'Без названия')}\n{r.get('content', '')}\nURL: {r.get('url', '')}"
            for i, r in enumerate(search_results["results"])
        ])
        
        # 3. Генерация ответа
        messages = [
            {
                "role": "system",
                "content": "Вы - помощник для поиска информации. Отвечайте кратко и точно, используя предоставленные источники. Указывайте номера источников в квадратных скобках [1], [2]."
            },
            {
                "role": "user",
                "content": f"Вопрос: {query}\n\nНайденная информация:\n{context}\n\nДайте краткий и точный ответ."
            }
        ]
        
        answer = await llm_client.chat_completion(
            messages=messages,
            temperature=0.3,
            max_tokens=500
        )
        
        return {
            "mode": "simple",
            "query": query,
            "answer": answer,
            "sources": search_results["results"],
            "response_time": time.time() - start_time
        }
        
    except Exception as e:
        logger.error(f"Simple mode error: {e}")
        return {
            "mode": "simple",
            "query": query,
            "answer": f"Произошла ошибка: {str(e)}",
            "sources": [],
            "response_time": time.time() - start_time
        }