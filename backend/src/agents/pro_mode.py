from typing import Dict, List
import time
import logging

logger = logging.getLogger(__name__)

async def process_pro_mode(
    query: str,
    search_client,
    llm_client,
    max_results: int = 10,
    context: str = None,
    previous_messages: List[Dict] = None
) -> Dict:
    start_time = time.time()
    reasoning_steps = []
    
    try:
        use_context = context and previous_messages and len(previous_messages) >= 2
        
        # Шаг 1: Анализ запроса с учетом контекста
        if use_context:
            reasoning_steps.append("Анализирую запрос с учётом контекста диалога...")
            
            messages = [{
                "role": "system",
                "content": "Проанализируйте запрос с учётом контекста диалога. Разбейте на 2-3 конкретных поисковых подзапроса. Отвечайте в формате: 1. запрос\n2. запрос\n3. запрос"
            }, {
                "role": "user",
                "content": f"Контекст диалога:\n{context}\n\nТекущий вопрос: {query}\n\nСоздайте поисковые подзапросы:"
            }]
        else:
            reasoning_steps.append("📋 Анализирую запрос и генерирую подзапросы...")
            
            messages = [{
                "role": "system",
                "content": "Разбейте запрос на 2-3 конкретных поисковых подзапроса. Отвечайте в формате: 1. запрос\n2. запрос\n3. запрос"
            }, {
                "role": "user",
                "content": f"Разбейте этот вопрос на подзапросы: {query}"
            }]
        
        subqueries_text = await llm_client.chat_completion(messages, temperature=0.5, max_tokens=200)
        subqueries = [q.strip() for q in subqueries_text.split('\n') if q.strip() and not q.strip().startswith('#')][:3]
        
        if not subqueries:
            subqueries = [query]
        
        reasoning_steps.append(f"Создано {len(subqueries)} подзапросов: {', '.join(subqueries)}")
        
        # Шаг 2: Множественный поиск
        all_results = []
        for subquery in subqueries:
            results = await search_client.search(
                query=subquery,
                max_results=max_results,
                include_raw_content=True
            )
            all_results.extend(results.get("results", []))
        
        reasoning_steps.append(f"📊 Получено {len(all_results)} результатов из поиска")
        
        if not all_results:
            return {
                "mode": "pro",
                "query": query,
                "answer": "Не удалось найти достаточно информации для анализа.",
                "sources": [],
                "reasoning_steps": reasoning_steps,
                "search_queries": subqueries,
                "response_time": time.time() - start_time,
                "context_used": False
            }
        
        # Шаг 3: Анализ фактов
        reasoning_steps.append("Анализирую и проверяю факты из источников...")
        
        search_context = "\n\n".join([
            f"Источник {i+1}: {r.get('title', '')}\n{r.get('raw_content', r.get('content', ''))[:1000]}\nURL: {r.get('url', '')}"
            for i, r in enumerate(all_results[:5])
        ])
        
        # Шаг 4: Синтез ответа с учётом контекста диалога
        reasoning_steps.append("Формирую итоговый ответ с цитированием источников...")
        
        if use_context:
            system_prompt = """Создайте подробный ответ на основе проверенной информации и контекста диалога.
Включите цитирование источников [1], [2] и структурируйте ответ.
Если вопрос связан с предыдущим диалогом, учитывайте этот контекст в своём ответе."""
            
            user_prompt = f"""Контекст диалога:
{context}

Текущий вопрос: {query}

Информация из источников:
{search_context}

Создайте полный ответ с анализом и выводами, учитывая контекст диалога."""
        else:
            system_prompt = "Создайте подробный ответ на основе проверенной информации. Включите цитирование источников [1], [2] и структурируйте ответ."
            user_prompt = f"Вопрос: {query}\n\nИнформация из источников:\n{search_context}\n\nСоздайте полный ответ с анализом и выводами."
        
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt}
        ]
        
        answer = await llm_client.chat_completion(messages, temperature=0.5, max_tokens=1500)
        
        reasoning_steps.append("✅ Ответ готов")
        
        return {
            "mode": "pro",
            "query": query,
            "answer": answer,
            "sources": all_results[:5],
            "reasoning_steps": reasoning_steps,
            "search_queries": subqueries,
            "response_time": time.time() - start_time,
            "context_used": use_context
        }
        
    except Exception as e:
        logger.error(f"Pro mode error: {e}")
        reasoning_steps.append(f"Ошибка: {str(e)}")
        return {
            "mode": "pro",
            "query": query,
            "answer": f"Произошла ошибка при обработке: {str(e)}",
            "sources": [],
            "reasoning_steps": reasoning_steps,
            "search_queries": [],
            "response_time": time.time() - start_time,
            "context_used": False
        }