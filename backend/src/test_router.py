import asyncio
import sys
sys.path.append("src")

from agents.router_agent import RouterAgent
from tools.llm_factory import create_llm_client
from core.config import settings

async def test_router():
    # Инициализация
    llm_client = create_llm_client(
        provider=settings.llm_provider,
        api_key=settings.openai_api_key if settings.llm_provider == "openai" else settings.llm_api_key,
        base_url=settings.llm_api_url if settings.llm_provider == "local" else None,
        model=settings.openai_model if settings.llm_provider == "openai" else None
    )
    
    router = RouterAgent(llm_client)
    
    # Тестовые запросы
    test_queries = [
        # Simple
        ("What is Python?", False),
        ("Когда основан Google?", False),
        ("Кто президент США?", False),
        
        # Pro
        ("Сравни Python и Java для веб-разработки", False),
        ("Почему биткоин растет в цене?", False),
        ("Проанализируй плюсы и минусы удаленной работы", False),
        
        # С контекстом
        ("Расскажи об этом подробнее", True),
        ("А что насчет производительности?", True),
    ]
    
    print("=" * 80)
    print("ТЕСТИРОВАНИЕ ROUTER AGENT")
    print("=" * 80)
    
    for query, has_context in test_queries:
        print(f"\n📝 Запрос: '{query}'")
        print(f"   Контекст: {'есть' if has_context else 'нет'}")
        
        # Только эвристика
        result_heuristic = await router.route(query, use_llm=False, context_exists=has_context)
        print(f"   🔧 Эвристика: {result_heuristic['mode'].upper()} "
              f"({result_heuristic['confidence']:.0%}) - {result_heuristic['reason']}")
        
        # С LLM
        if settings.use_llm_router:
            result_llm = await router.route(query, use_llm=True, context_exists=has_context)
            print(f"   🤖 LLM:       {result_llm['mode'].upper()} "
                  f"({result_llm['confidence']:.0%}) - {result_llm['reason']}")
    
    await llm_client.close()
    print("\n" + "=" * 80)
    print("✅ Тестирование завершено")

if __name__ == "__main__":
    asyncio.run(test_router())