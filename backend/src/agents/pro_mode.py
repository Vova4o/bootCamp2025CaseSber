from typing import Dict, List
import logging
import time

logger = logging.getLogger(__name__)


async def process_pro_mode_with_context(
    query: str,
    search_client,
    llm_client,
    conversation_history: List[Dict] = None,
    max_results: int = 10
) -> Dict:
    """
    Pro Mode с учётом контекста предыдущих сообщений.
    
    Args:
        query: Текущий запрос
        search_client: Клиент для поиска
        llm_client: Клиент LLM
        conversation_history: История диалога [{"role": "user/assistant", "content": "..."}]
        max_results: Максимум результатов поиска
    """
    start_time = time.time()
    reasoning_steps = []
    
    try:
        # Шаг 1: Анализ контекста и формирование улучшенного запроса
        if conversation_history and len(conversation_history) > 0:
            reasoning_steps.append("🔍 Анализирую контекст предыдущего диалога...")
            
            # Формируем контекст для LLM
            context_prompt = "Предыдущая беседа:\n"
            for msg in conversation_history[-6:]:  # Последние 3 пары вопрос-ответ
                role = "Пользователь" if msg["role"] == "user" else "Ассистент"
                context_prompt += f"\n{role}: {msg['content']}\n"
            
            # Улучшаем запрос с учётом контекста
            enhanced_query = await llm_client.chat_completion(
                messages=[
                    {
                        "role": "system",
                        "content": "Ты помощник, который улучшает поисковые запросы с учётом контекста диалога. Перефразируй текущий вопрос так, чтобы он был самодостаточным и включал важную информацию из контекста."
                    },
                    {
                        "role": "user",
                        "content": f"{context_prompt}\n\nТекущий вопрос: {query}\n\nУлучшенный поисковый запрос:"
                    }
                ],
                temperature=0.3,
                max_tokens=200
            )
            
            reasoning_steps.append(f"✨ Улучшенный запрос: {enhanced_query}")
            search_query = enhanced_query
        else:
            search_query = query
            reasoning_steps.append("📝 Обрабатываю первый запрос без контекста")
        
        # Шаг 2: Поиск информации
        reasoning_steps.append(f"🔎 Ищу информацию по запросу: {search_query}")
        search_results = await search_client.search(
            query=search_query,
            max_results=max_results,
            include_raw_content=True
        )
        
        if not search_results.get("results"):
            return {
                "query": query,
                "mode": "pro",
                "answer": "Не удалось найти релевантную информацию.",
                "sources": [],
                "reasoning": "\n".join(reasoning_steps),
                "processing_time": time.time() - start_time,
                "timestamp": time.time(),
                "context_used": len(conversation_history) > 0 if conversation_history else False
            }
        
        reasoning_steps.append(f"✅ Найдено {len(search_results['results'])} источников")
        
        # Шаг 3: Формирование ответа с учётом контекста
        sources_context = "\n\n".join([
            f"Источник {i+1} ({r.get('title', 'Unknown')}):\n{r.get('content', r.get('snippet', ''))}"
            for i, r in enumerate(search_results["results"][:5])
        ])
        
        # Собираем messages для LLM
        llm_messages = [
            {
                "role": "system",
                "content": """Ты исследовательский ассистент в режиме Pro. 
                Твоя задача - дать подробный, хорошо обоснованный ответ с учётом:
                1. Контекста предыдущей беседы
                2. Найденных источников
                3. Проверки фактов
                
                Формат ответа:
                - Прямой ответ на вопрос
                - Подтверждение фактами из источников
                - Цитирование источников
                - Если информация противоречива - укажи это"""
            }
        ]
        
        # Добавляем контекст диалога
        if conversation_history:
            for msg in conversation_history[-4:]:  # Последние 2 пары
                llm_messages.append({
                    "role": msg["role"],
                    "content": msg["content"]
                })
        
        # Добавляем текущий вопрос и источники
        llm_messages.append({
            "role": "user",
            "content": f"Вопрос: {query}\n\nНайденная информация:\n{sources_context}\n\nОтвет:"
        })
        
        reasoning_steps.append("💡 Формирую ответ с учётом всех данных...")
        
        answer = await llm_client.chat_completion(
            messages=llm_messages,
            temperature=0.7,
            max_tokens=1000
        )
        
        # Форматируем источники
        formatted_sources = [
            {
                "title": r.get("title", "Unknown"),
                "url": r.get("url", "#"),
                "snippet": r.get("snippet", "")[:200],
                "credibility": 0.85  # TODO: реальная оценка достоверности
            }
            for r in search_results["results"][:5]
        ]
        
        return {
            "query": query,
            "mode": "pro",
            "answer": answer,
            "sources": formatted_sources,
            "reasoning": "\n".join(reasoning_steps),
            "processing_time": time.time() - start_time,
            "timestamp": time.time(),
            "context_used": len(conversation_history) > 0 if conversation_history else False
        }
        
    except Exception as e:
        logger.error(f"Pro mode with context error: {e}")
        return {
            "query": query,
            "mode": "pro",
            "answer": f"Ошибка обработки: {str(e)}",
            "sources": [],
            "reasoning": "\n".join(reasoning_steps + [f"❌ Ошибка: {str(e)}"]),
            "processing_time": time.time() - start_time,
            "timestamp": time.time(),
            "context_used": False
        }