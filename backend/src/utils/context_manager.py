from typing import List, Dict
import logging

logger = logging.getLogger(__name__)

class ContextManager:
    """
    Управление контекстным окном для Pro Mode
    Simple Mode не использует контекст - работает только с текущим запросом
    """
    
    def __init__(self, max_messages: int = 10, max_tokens: int = 4000):
        self.max_messages = max_messages
        self.max_tokens = max_tokens
    
    def build_context(self, messages: List[Dict]) -> str:
        """
        Создает текстовый контекст из истории сообщений для Pro Mode
        
        Args:
            messages: Список сообщений [{role, content, timestamp}]
        
        Returns:
            Форматированный контекст
        """
        if not messages:
            return ""
        
        # Берем последние N сообщений
        recent_messages = messages[-self.max_messages:]
        
        context_parts = ["Контекст предыдущего разговора:"]
        for msg in recent_messages:
            role = "Пользователь" if msg["role"] == "user" else "Ассистент"
            context_parts.append(f"{role}: {msg['content']}")
        
        context = "\n".join(context_parts)
        
        # Обрезаем по токенам (примерно 4 символа = 1 токен)
        max_chars = self.max_tokens * 4
        if len(context) > max_chars:
            context = context[-max_chars:]
            context = "...\n" + context.split("\n", 1)[1]  # Убираем неполное первое сообщение
        
        return context
    
    def should_use_context(self, query: str, messages: List[Dict]) -> bool:
        """
        Определяет, нужен ли контекст для ответа в Pro Mode
        
        Args:
            query: Текущий запрос
            messages: История сообщений
        
        Returns:
            True если нужен контекст
        """
        if not messages or len(messages) < 2:
            return False
        
        # Ключевые слова, указывающие на использование контекста
        context_indicators = [
            "это", "этого", "этом", "этой", "этому",
            "он", "она", "оно", "они",
            "тот", "та", "то", "те",
            "такой", "такая", "такое", "такие",
            "его", "её", "их",
            "также", "еще", "ещё", "тоже",
            "а как", "а что",
            "продолжи", "расскажи больше",
            "подробнее", "детальнее",
        ]
        
        query_lower = query.lower()
        return any(indicator in query_lower for indicator in context_indicators)
    
    def format_messages_for_llm(self, messages: List[Dict], new_query: str) -> List[Dict[str, str]]:
        """
        Форматирует сообщения для отправки в LLM (только для Pro Mode)
        
        Args:
            messages: История сообщений
            new_query: Новый запрос пользователя
        
        Returns:
            Список сообщений в формате LLM
        """
        # Берем последние N сообщений
        recent_messages = messages[-self.max_messages:] if messages else []
        
        llm_messages = []
        
        # Добавляем контекст предыдущих сообщений
        for msg in recent_messages:
            llm_messages.append({
                "role": msg["role"],
                "content": msg["content"]
            })
        
        # Добавляем новый запрос
        llm_messages.append({
            "role": "user",
            "content": new_query
        })
        
        return llm_messages