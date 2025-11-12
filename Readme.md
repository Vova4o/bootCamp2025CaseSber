# Research Pro Mode - Sber Bootcamp 2025

Продвинутый поисковый ассистент с двумя режимами работы: Simple Mode для быстрого поиска и Pro Mode для глубокого исследования с проверкой фактов.

## 🎯 Описание проекта

Research Pro Mode — это мультиагентная система, которая не просто ищет ответы, а понимает контекст, сравнивает источники и проверяет факты. Система использует несколько специализированных агентов для обработки запросов пользователя.

### Режимы работы

#### Simple Mode
- Быстрый поиск с использованием SERP API
- Минимальные накладные расходы
- Идеален для простых запросов
- Быстрое время отклика

#### Pro Mode
- Глубокий веб-скрапинг
- Семантическая переоценка результатов
- Многошаговое рассуждение (multi-hop reasoning)
- Проверка фактов из нескольких источников
- Детальный анализ с цитированием источников

## 🏗️ Архитектура системы

```
┌─────────────────┐
│  User Request   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│  1. Mode Selector Agent │ ◄── Выбор Simple/Pro режима
└────────┬────────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌───────┐  ┌──────────────────────┐
│Simple │  │    Pro Mode Flow     │
│ Mode  │  └──────────┬───────────┘
└───┬───┘             │
    │            ┌────┴────┐
    │            ▼         │
    │  ┌──────────────────────────┐
    │  │ 2. Web Scraping Agent    │ ◄── Сбор данных
    │  └──────────┬───────────────┘
    │             │
    │             ▼
    │  ┌──────────────────────────┐
    │  │ 3. Fact Analysis Agent   │ ◄── Анализ фактов
    │  └──────────┬───────────────┘
    │             │
    │             ▼
    │  ┌──────────────────────────┐
    │  │ 4. Hypothesis Verification│ ◄── Проверка гипотез
    │  └──────────┬───────────────┘
    │             │
    └────┬────────┘
         │
         ▼
┌─────────────────────────┐
│ 5. Response Aggregator  │ ◄── Сбор итогового ответа
└────────┬────────────────┘
         │
         ▼
┌─────────────────┐
│  6. User Reply  │
└─────────────────┘
```

## 🤖 Агенты системы

### 1. Mode Selector Agent
**Задача**: Анализирует запрос пользователя и выбирает оптимальный режим работы.

**Критерии выбора**:
- Simple Mode: простые фактические вопросы, актуальная информация
- Pro Mode: сложные запросы, требующие анализа, сравнения или проверки

### 2. Web Scraping Agent
**Задача**: Сбор информации из веб-источников.

**Инструменты**:
- SerpAPI / Brave Search API для поисковых запросов
- BeautifulSoup / Playwright для скрапинга контента
- searxng-docker-tavily-adapter для агрегации результатов

### 3. Fact Analysis Agent
**Задача**: Анализ собранных фактов и выделение ключевой информации.

**Функции**:
- Семантический анализ текста
- Выделение ключевых фактов
- Оценка релевантности источников

### 4. Hypothesis Verification Agent
**Задача**: Проверка гипотез и перекрёстная проверка фактов.

**Функции**:
- Сравнение данных из разных источников
- Выявление противоречий
- Оценка достоверности информации

### 5. Response Aggregator
**Задача**: Формирование итогового ответа с цитированием источников.

**Функции**:
- Структурирование информации
- Добавление ссылок на источники
- Объяснение логики рассуждений

### 6. User Interface
**Задача**: Взаимодействие с пользователем и отправка ответа.

## 📁 Структура проекта

```
bootCamp2025CaseSber/
├── src/
│   ├── agents/
│   │   ├── __init__.py
│   │   ├── mode_selector.py       # Агент выбора режима
│   │   ├── web_scraper.py         # Агент веб-скрапинга
│   │   ├── fact_analyzer.py       # Агент анализа фактов
│   │   ├── hypothesis_verifier.py # Агент проверки гипотез
│   │   └── response_aggregator.py # Агент сборки ответа
│   ├── core/
│   │   ├── __init__.py
│   │   ├── orchestrator.py        # Оркестратор агентов
│   │   └── config.py              # Конфигурация
│   ├── tools/
│   │   ├── __init__.py
│   │   ├── search_api.py          # Интеграция с SERP API
│   │   ├── scraper.py             # Инструменты скрапинга
│   │   └── llm_client.py          # Клиент для LLM
│   └── utils/
│       ├── __init__.py
│       ├── logger.py
│       └── helpers.py
├── tests/
│   ├── test_agents/
│   ├── test_tools/
│   └── benchmarks/
│       ├── simpleqa_bench.py
│       └── frames_bench.py
├── notebooks/
│   └── demo.ipynb
├── docs/
│   ├── architecture.md
│   ├── api_docs.md
│   └── benchmarks.md
├── .env.example
├── .gitignore
├── requirements.txt
└── README.md
```

## 🚀 Установка и запуск

### Требования
- Python 3.10+
- API ключи для поисковых сервисов (SerpAPI/Brave Search)
- LLM API (OpenAI/Anthropic или локальная модель)

### Установка

```bash
# Клонирование репозитория
git clone <repository-url>
cd bootCamp2025CaseSber

# Создание виртуального окружения
python -m venv venv
source venv/bin/activate  # для Mac/Linux

# Установка зависимостей
pip install -r requirements.txt

# Настройка переменных окружения
cp .env.example .env
# Отредактируйте .env и добавьте ваши API ключи
```

### Конфигурация

```bash
# .env файл
SERP_API_KEY=your_serp_api_key
BRAVE_API_KEY=your_brave_api_key
OPENAI_API_KEY=your_openai_key
LLM_MODEL=gpt-4
SEARCH_ENGINE=searxng  # или brave, serp
```

## 💻 Использование

### Простой пример

```python
from src.core.orchestrator import ResearchOrchestrator

# Инициализация системы
orchestrator = ResearchOrchestrator()

# Simple Mode
response = orchestrator.process_query(
    query="Какая столица Франции?",
    mode="auto"  # Автоматический выбор режима
)

# Pro Mode
response = orchestrator.process_query(
    query="Сравните подходы к регулированию AI в США, ЕС и Китае",
    mode="pro"
)

print(response)
```

### CLI интерфейс

```bash
# Simple Mode
python -m src.main "Какая столица Франции?"

# Принудительный Pro Mode
python -m src.main "Сравните подходы к регулированию AI" --mode pro

# С детальным логированием
python -m src.main "Ваш запрос" --verbose
```

## 📊 Бенчмарки

### SimpleQA Bench
Тестирование точности на простых фактических вопросах.

```bash
python -m tests.benchmarks.simpleqa_bench --mode simple
```

**Метрика**: Accuracy (%)

### FRAMES Bench
Оценка многошагового рассуждения и работы с источниками.

```bash
python -m tests.benchmarks.frames_bench --mode pro
```

**Метрики**:
- Factuality (точность фактов)
- Reasoning Depth (глубина рассуждений)
- Source Diversity (разнообразие источников)

## 🔧 Расширенные режимы (опционально)

### Pro: Social
Анализ мнений из социальных сетей (Reddit, X, VK, Habr).

### Pro: Academic
Поиск в научных базах (arXiv, Semantic Scholar).

### Pro: Finance
Финансовые данные (Yahoo Finance, TradingView).

## 🧪 Тестирование

```bash
# Запуск всех тестов
pytest tests/

# Тесты агентов
pytest tests/test_agents/

# Бенчмарки
pytest tests/benchmarks/
```

## 📝 Примеры запросов

### Simple Mode
- "Когда был основан Google?"
- "Какая погода в Москве?"
- "Кто президент США?"

### Pro Mode
- "Сравните различные подходы к квантовым вычислениям"
- "Проанализируйте влияние AI на рынок труда за последние 5 лет"
- "Какие существуют теории о происхождении вселенной и их доказательства?"

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

MIT License

## 👥 Команда

Sber Bootcamp 2025 - Case Study

## 🔗 Полезные ссылки

- [SearXNG Docker Tavily Adapter](https://github.com/vakovalskii/searxng-docker-tavily-adapter)
- [SimpleQA Benchmark](https://openai.com/index/introducing-simpleqa/)
- [FRAMES Benchmark](https://arxiv.org/abs/2409.12941)
- [LangChain Documentation](https://python.langchain.com/)
- [LangGraph Documentation](https://langchain-ai.github.io/langgraph/)

## 📈 Roadmap

- [x] Базовая архитектура мультиагентной системы
- [ ] Реализация Simple Mode
- [ ] Реализация Pro Mode
- [ ] Интеграция с поисковыми API
- [ ] Веб-скрапинг агент
- [ ] Система проверки фактов
- [ ] Бенчмарки SimpleQA и FRAMES
- [ ] Web UI интерфейс
- [ ] Расширенные режимы (Social, Academic, Finance)
- [ ] Docker контейнеризация
- [ ] CI/CD pipeline