from typing import Dict, Literal
import logging
import re

logger = logging.getLogger(__name__)

class RouterAgent:
    """
    Агент-маршрутизатор для определения сложности запроса
    и выбора между Simple и Pro режимами
    """
    
    def __init__(self, llm_client):
        self.llm_client = llm_client
        
        # Ключевые слова для быстрой эвристики
        self.pro_keywords = [
            # Сравнения и анализ
            "сравни", "сравнить", "compare", "отличие", "difference", "versus", "vs",
            "лучше", "хуже", "better", "worse",
            
            # Множественные аспекты
            "плюсы и минусы", "преимущества и недостатки", "pros and cons",
            "advantages", "disadvantages",
            
            # Глубокий анализ
            "проанализируй", "analyze", "исследуй", "research",
            "подробно", "детально", "detailed", "comprehensive",
            "почему", "why", "как работает", "how does", "how works",
            
            # Проверка фактов
            "правда ли", "is it true", "факт", "fact check",
            "достоверно", "reliable", "проверь", "verify",
            
            # Множественные источники
            "по данным", "according to", "исследования показывают",
            "эксперты", "experts", "мнения", "opinions",
            
            # Сложные вопросы
            "объясни", "explain", "расскажи подробно", "tell me more",
            "каким образом", "how exactly", "в чем причина", "what causes",
        ]
        
        self.simple_keywords = [
            # Простые определения
            "что такое", "what is", "кто такой", "who is",
            "определение", "definition", "значение", "meaning",
            
            # Простые факты
            "когда", "when", "где", "where", "сколько", "how many", "how much",
            "дата", "date", "год", "year",
            
            # Простые действия
            "как сделать", "how to", "инструкция", "instruction",
        ]
    
    async def route(
        self, 
        query: str, 
        use_llm: bool = True,
        context_exists: bool = False
    ) -> Dict[Literal["mode", "confidence", "reason"], any]:
        """
        Определяет режим обработки запроса
        
        Args:
            query: Запрос пользователя
            use_llm: Использовать ли LLM для классификации
            context_exists: Есть ли контекст диалога
        
        Returns:
            {
                "mode": "simple" | "pro",
                "confidence": float (0-1),
                "reason": str
            }
        """
        # Шаг 1: Быстрая эвристика
        heuristic_result = self._heuristic_check(query, context_exists)
        
        # Если уверенность высокая (>0.8), возвращаем результат эвристики
        if heuristic_result["confidence"] > 0.8 or not use_llm:
            logger.info(
                f"Router decision (heuristic): {heuristic_result['mode']} "
                f"(confidence: {heuristic_result['confidence']:.2f}) - {heuristic_result['reason']}"
            )
            return heuristic_result
        
        # Шаг 2: LLM классификация для неоднозначных случаев
        try:
            llm_result = await self._llm_classify(query, context_exists)
            logger.info(
                f"Router decision (LLM): {llm_result['mode']} "
                f"(confidence: {llm_result['confidence']:.2f}) - {llm_result['reason']}"
            )
            return llm_result
        except Exception as e:
            logger.error(f"LLM classification failed: {e}, using heuristic")
            return heuristic_result
    
    def _heuristic_check(self, query: str, context_exists: bool) -> Dict:
        """Быстрая эвристическая проверка"""
        query_lower = query.lower()
        words = query.split()
        
        # Критерий 1: Длина запроса
        if len(words) <= 4:
            return {
                "mode": "simple",
                "confidence": 0.9,
                "reason": "Короткий запрос (≤4 слов)"
            }
        
        if len(words) >= 15:
            return {
                "mode": "pro",
                "confidence": 0.85,
                "reason": "Длинный сложный запрос (≥15 слов)"
            }
        
        # Критерий 2: Наличие контекста
        if context_exists:
            return {
                "mode": "pro",
                "confidence": 0.8,
                "reason": "Есть контекст диалога - используем Pro для связности"
            }
        
        # Критерий 3: Ключевые слова Pro режима
        pro_matches = sum(1 for keyword in self.pro_keywords if keyword in query_lower)
        if pro_matches >= 2:
            return {
                "mode": "pro",
                "confidence": 0.9,
                "reason": f"Найдено {pro_matches} маркеров сложности"
            }
        
        # Критерий 4: Ключевые слова Simple режима
        simple_matches = sum(1 for keyword in self.simple_keywords if keyword in query_lower)
        if simple_matches >= 1:
            return {
                "mode": "simple",
                "confidence": 0.85,
                "reason": f"Найдено {simple_matches} маркеров простого запроса"
            }
        
        # Критерий 5: Вопросительные слова
        question_words = ["как", "что", "где", "когда", "почему", "зачем", 
                         "how", "what", "where", "when", "why"]
        if any(query_lower.startswith(q) for q in question_words):
            # "Как работает" - pro, "Как называется" - simple
            if any(word in query_lower for word in ["работает", "функционирует", "устроен", "works", "functions"]):
                return {
                    "mode": "pro",
                    "confidence": 0.75,
                    "reason": "Вопрос о механизме работы"
                }
        
        # Критерий 6: Множественные вопросы
        if query.count("?") > 1 or query.count(" и ") > 2:
            return {
                "mode": "pro",
                "confidence": 0.8,
                "reason": "Множественные вопросы или аспекты"
            }
        
        # Критерий 7: По умолчанию - средняя сложность
        if len(words) <= 8:
            return {
                "mode": "simple",
                "confidence": 0.6,
                "reason": "Средняя сложность, склоняемся к Simple для скорости"
            }
        else:
            return {
                "mode": "pro",
                "confidence": 0.6,
                "reason": "Средняя сложность, склоняемся к Pro для качества"
            }
    
    async def _llm_classify(self, query: str, context_exists: bool) -> Dict:
        """LLM классификация для неоднозначных случаев"""
        
        system_prompt = """Ты - агент-классификатор запросов. Определи сложность вопроса.

SIMPLE MODE (⚡ быстрый):
- Простые фактические вопросы
- Определения и значения
- Даты, числа, простые факты
- Короткие ответы из одного источника
Примеры: "Что такое Python?", "Когда основан Google?", "Кто автор книги X?"

PRO MODE (🧠 глубокий):
- Сравнения и анализ
- Множественные аспекты
- Проверка фактов
- Вопросы требующие рассуждений
- Сложные "почему" и "как работает"
Примеры: "Сравни Python и Java", "Почему Bitcoin растет?", "Как работает нейросеть?"

Отвечай ТОЛЬКО в формате: MODE|CONFIDENCE|REASON
Где MODE = simple или pro, CONFIDENCE = 0.0-1.0, REASON = краткое объяснение

Пример: pro|0.85|Требует анализа и сравнения"""

        user_prompt = f"""Запрос: "{query}"
Есть контекст диалога: {"да" if context_exists else "нет"}

Классифицируй запрос:"""

        try:
            messages = [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt}
            ]
            
            response = await self.llm_client.chat_completion(
                messages=messages,
                temperature=0.3,
                max_tokens=100
            )
            
            # Парсинг ответа
            parts = response.strip().split("|")
            if len(parts) >= 3:
                mode = parts[0].strip().lower()
                confidence = float(parts[1].strip())
                reason = parts[2].strip()
                
                if mode not in ["simple", "pro"]:
                    mode = "simple"
                
                confidence = max(0.0, min(1.0, confidence))
                
                return {
                    "mode": mode,
                    "confidence": confidence,
                    "reason": f"LLM: {reason}"
                }
            else:
                raise ValueError("Invalid response format")
                
        except Exception as e:
            logger.error(f"LLM parsing error: {e}, response: {response if 'response' in locals() else 'N/A'}")
            # Fallback к эвристике
            return self._heuristic_check(query, context_exists)