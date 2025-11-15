# Research Pro Mode - Sber Bootcamp 2025

Продвинутый поисковый ассистент с тремя режимами работы и Telegram ботом. Система обеспечивает быстрый поиск (Simple), глубокий анализ с контекстом (Pro) и автоматический выбор режима (Auto).

## 🎯 Описание проекта

Research Pro Mode — это мультиагентная система на Go с веб-интерфейсом и Telegram ботом. Система использует DuckDuckGo для поиска, LLM (OpenAI/Qwen) для анализа и поддерживает контекст беседы в Pro режиме.

**Технологический стек:**

- **Backend**: Go 1.23+ с Gin web framework
- **Frontend**: Next.js 14+ с TypeScript
- **Telegram Bot**: go-telegram-bot-api/v5
- **Database**: PostgreSQL + Redis
- **Search**: DuckDuckGo (HTML scraping + Instant Answer API)
- **LLM**: OpenAI GPT-4 / Qwen API
- **Архитектура**: Monorepo с Docker Compose

### Режимы работы

#### 🤖 Auto Mode

- Автоматический выбор между Simple и Pro
- LLM анализирует сложность запроса
- Оптимизирует скорость и качество

#### ⚡ Simple Mode

- Быстрый поиск через DuckDuckGo
- Синтез результатов с помощью LLM
- Без сохранения контекста
- Идеален для простых вопросов (Кто? Что? Когда?)
- ~2-3 секунды на запрос

#### 🚀 Pro Mode

- Глубокий анализ с контекстом беседы
- Улучшение запроса на основе истории
- Многошаговое рассуждение
- Оценка достоверности источников
- Поддержка follow-up вопросов
- ~5-10 секунд на запрос

## 🏗️ Архитектура системы

```
┌──────────────────┐         ┌──────────────────┐
│   Next.js UI     │         │  Telegram Bot    │
│   (Port 3000)    │         │   (@bot)         │
└────────┬─────────┘         └────────┬─────────┘
         │                            │
         │         HTTP API           │
         └─────────────┬──────────────┘
                       ▼
              ┌─────────────────┐
              │   Gin Server    │ ◄── Go Backend (Port 8000)
              │   (Port 8000)   │
              └────────┬────────┘
                       │
         ┌─────────────┴─────────────┐
         ▼                           ▼
┌─────────────────┐         ┌─────────────────┐
│  RouterAgent    │         │  ModeSelector   │
│  (Auto/Simple/  │         │  (LLM-based)    │
│   Pro routing)  │         │                 │
└────────┬────────┘         └─────────────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌──────────────┐  ┌──────────────────┐
│ SimpleAgent  │  │   ProAgent       │
│              │  │                  │
│ • DuckDuckGo │  │ • Context from   │
│ • LLM Synth  │  │   chat history   │
│ • No context │  │ • Query enhance  │
└──────┬───────┘  │ • DuckDuckGo     │
       │          │ • Credibility    │
       │          │ • LLM reasoning  │
       │          └────────┬─────────┘
       └──────────┬────────┘
                  ▼
         ┌─────────────────┐
         │  SearchClient   │ ◄── DuckDuckGo scraping
         │  (Rate limited  │     + Instant Answer API
         │   1 req/sec)    │
         └────────┬────────┘
                  │
         ┌────────┴────────┐
         ▼                 ▼
┌─────────────────┐  ┌─────────────────┐
│   PostgreSQL    │  │     Redis       │
│  (Sessions &    │  │   (Caching)     │
│   Messages)     │  │                 │
└─────────────────┘  └─────────────────┘
```

## 📁 Структура проекта (Monorepo)

```
bootCamp2025CaseSber/
├── backend-go/                     # Go Backend
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go            # Gin HTTP server
│   │   ├── tgbot/
│   │   │   ├── main.go            # Telegram bot
│   │   │   └── README.md
│   │   └── benchmark/
│   │       ├── simpleqa/main.go   # SimpleQA benchmark
│   │       ├── frames/main.go     # FRAMES benchmark
│   │       └── compare/main.go    # Comparison tool
│   ├── internal/
│   │   ├── agents/
│   │   │   ├── router.go          # Agent routing
│   │   │   ├── simple_agent.go    # Fast search
│   │   │   ├── pro_agent.go       # Deep analysis
│   │   │   └── mode_selector.go   # Auto mode
│   │   ├── api/
│   │   │   ├── routes.go
│   │   │   └── handlers/
│   │   │       ├── chat.go        # Chat sessions
│   │   │       ├── search.go      # One-time search
│   │   │       └── health.go      # Health check
│   │   ├── tools/
│   │   │   ├── llm_client.go      # OpenAI/Qwen client
│   │   │   ├── search_client.go   # DuckDuckGo scraper
│   │   │   ├── reranker.go        # Result reranking
│   │   │   └── credibility.go     # Source scoring
│   │   ├── database/
│   │   │   └── database.go        # GORM models
│   │   ├── config/
│   │   │   └── config.go          # Config loader
│   │   └── models/
│   │       └── models.go          # API structs
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile                 # Backend image
│   └── Dockerfile.tgbot           # Telegram bot image
│
├── frontend/                       # Next.js Frontend
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx               # Main UI
│   │   └── globals.css
│   ├── components/
│   │   ├── ChatInterface.tsx      # Chat with context
│   │   ├── ChatMessage.tsx        # Message bubble
│   │   ├── ChatList.tsx           # Session list
│   │   ├── ModeSelector.tsx       # Mode switcher
│   │   └── CompactModeSelector.tsx
│   ├── lib/
│   │   └── api.ts                 # Axios client
│   ├── types/
│   │   └── index.ts               # TypeScript types
│   ├── package.json
│   ├── tsconfig.json
│   └── Dockerfile
│
├── searxng/                        # SearXNG config
│   ├── settings.yml
│   └── limiter.toml
│
├── docker-compose.yml              # Full stack
├── .env                            # Environment vars
├── .env.example                    # Template
├── .gitignore
└── README.md
```

## 🚀 Установка и запуск

### Требования

- Docker & Docker Compose
- Go 1.23+ (для локальной разработки)
- Node.js 18+ (для локальной разработки)
- OpenAI API ключ или Qwen API

### Quick Start с Docker (рекомендуется)

```bash
# 1. Клонировать репозиторий
git clone <repo-url>
cd bootCamp2025CaseSber

# 2. Настроить переменные окружения
cp .env.example .env
# Отредактируйте .env и добавьте:
# - OPENAI_API_KEY=your-key
# - TELEGRAM_BOT_TOKEN=your-bot-token (опционально)

# 3. Запустить весь стек
docker-compose up -d

# Проверить статус
docker-compose ps

# Логи
docker logs research_backend -f
docker logs research_frontend -f
docker logs research_tgbot -f
```

**Доступ:**

- Frontend: http://localhost:3000
- Backend API: http://localhost:8000
- SearXNG: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### Локальная разработка

#### Backend (Go)

```bash
cd backend-go

# Установить зависимости
go mod download

# Запустить сервер
go run ./cmd/server/main.go

# Или собрать
go build -o bin/server ./cmd/server/main.go
./bin/server
```

#### Frontend (Next.js)

```bash
cd frontend

# Установить зависимости
npm install

# Dev сервер
npm run dev

# Production build
npm run build
npm start
```

#### Telegram Bot

```bash
cd backend-go

# Получить токен от @BotFather в Telegram
# Добавить в .env: TELEGRAM_BOT_TOKEN=your-token

# Запустить бота
go run ./cmd/tgbot/main.go

# Или через Docker
docker-compose up -d tgbot
```

## 🎨 Frontend Features

- **Современный UI**: Использование shadcn/ui и Tailwind CSS
- **Real-time Updates**: Streaming ответов от агентов
- **Mode Switcher**: Переключение между Simple и Pro режимами
- **Source Citations**: Отображение источников с ссылками
- **Responsive Design**: Адаптивный дизайн для всех устройств
- **Dark Mode**: Поддержка тёмной темы

## 🔧 API Endpoints (скоро)

## 📋 API Endpoints

### Health Check

```bash
GET /health

# Response: {"status": "ok"}
```

### Stateless Search (без контекста)

```bash
POST /api/search
Content-Type: application/json

{
  "query": "quantum computing applications",
  "mode": "simple"  # "simple", "pro", "auto"
}
```

**Response:**

```json
{
  "query": "quantum computing applications",
  "answer": "Квантовые вычисления применяются в области криптографии, оптимизации...",
  "sources": [
    {
      "title": "Quantum Computing in Drug Discovery",
      "url": "https://example.com/article",
      "snippet": "Recent advances show...",
      "credibility_score": 0.95
    }
  ],
  "mode": "simple",
  "processing_time": "2.3s"
}
```

### Stateful Chat (с контекстом)

**1. Создать сессию:**

```bash
POST /api/chat/session
Content-Type: application/json

{
  "mode": "pro"  # "simple", "pro", "auto"
}

# Response: {"session_id": "550e8400-e29b-41d4-a716-446655440000"}
```

**2. Отправить сообщение:**

```bash
POST /api/chat/session/{session_id}/message
Content-Type: application/json

{
  "query": "Who is Donald Trump?",
  "mode": "pro"
}

# Второе сообщение с контекстом:
POST /api/chat/session/{session_id}/message
{
  "query": "How old is he?",
  "mode": "pro"
}
# Ответ использует контекст предыдущего сообщения
```

**3. Получить историю сессии:**

```bash
GET /api/chat/session/{session_id}/history

# Response:
{
  "session_id": "...",
  "messages": [
    {"role": "user", "content": "Who is Donald Trump?"},
    {"role": "assistant", "content": "Donald Trump is...", "sources": [...]},
    {"role": "user", "content": "How old is he?"},
    {"role": "assistant", "content": "He is 78 years old...", "sources": [...]}
  ]
}
```

**4. Удалить сессию:**

```bash
DELETE /api/chat/session/{session_id}
```

## 📊 Бенчмарки

### SimpleQA Benchmark

Тестирование на датасете SimpleQA (4,326 вопросов) - оценка точности фактических ответов:

```bash
cd backend-go

# Запустить SimpleQA бенчмарк
make benchmark-simpleqa
# или
./bin/simpleqa

# Результаты в: cmd/benchmark/simpleqa/simpleqa_benchmark_*.json
```

**Метрики:**

- Accuracy: % правильных ответов
- Response Time: среднее время обработки
- Source Quality: средний балл достоверности источников
- Citation Rate: % ответов с источниками

### FRAMES Benchmark

Тестирование на FRAMES (824 вопроса) - оценка рассуждений с реальными фактами:

```bash
cd backend-go

# Запустить FRAMES бенчмарк
make benchmark-frames
# или
go run ./cmd/benchmark/frames/main.go

# Результаты в: cmd/benchmark/frames/frames_benchmark_*.json
```

**Метрики:**

- Reasoning Accuracy: корректность логических цепочек
- Fact Retrieval: точность извлечения фактов
- Multi-hop Performance: качество многоступенчатых выводов

### Сравнение режимов

```bash
cd backend-go

# Сравнить Simple vs Pro режимы
./bin/compare simpleqa_benchmark_1.json simpleqa_benchmark_2.json
# или
go run ./cmd/benchmark/compare/main.go file1.json file2.json
```

**Типичные результаты:**

- Simple Mode: ~2-3s, 70-75% accuracy
- Pro Mode: ~5-10s, 80-85% accuracy (с контекстом)
- Auto Mode: динамический выбор на основе сложности запроса

## 🧪 Тестирование

### Backend Tests (Go)

```bash
cd backend-go

# Запустить все тесты
go test ./...

# С покрытием
go test -cover ./...

# Verbose режим
go test -v ./internal/agents/

# Конкретный пакет
go test ./internal/tools/
```

### Frontend Tests

```bash
cd frontend

# Unit тесты
npm test

# E2E тесты
npm run test:e2e

# Coverage
npm run test:coverage
```

### Integration Tests

```bash
# Запустить Docker stack
docker-compose up -d

# Тестовый запрос к API
curl -X POST http://localhost:8000/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "What is AI?", "mode": "simple"}'

# Health check
curl http://localhost:8000/health
```

## 📈 Roadmap

### ✅ Completed (v1.0)

- [x] Базовая архитектура мультиагентной системы (Go)
- [x] Monorepo структура с Docker Compose
- [x] Next.js frontend с TypeScript
- [x] Gin backend с REST API
- [x] RouterAgent с ModeSelector (LLM-based)
- [x] Simple Mode (быстрый поиск 2-3s)
- [x] Pro Mode (контекстный поиск 5-10s)
- [x] Auto Mode (автоопределение сложности)
- [x] Интеграция DuckDuckGo (HTML + Instant API)
- [x] Система credibility scoring источников
- [x] PostgreSQL + Redis для персистентности
- [x] Telegram Bot интеграция (@aiassistanthelp_bot)
- [x] Бенчмарки SimpleQA и FRAMES
- [x] Docker контейнеризация всего стека
- [x] SearXNG метапоисковая система

### 🔄 In Progress

- [ ] Context preservation в Telegram боте (chat sessions)
- [ ] Улучшенный reranker с кроссэнкодером
- [ ] Расширенные специализированные агенты:
  - [ ] Social Media Agent (Reddit, Twitter scraping)
  - [ ] Academic Agent (arXiv, Scholar, PubMed)
  - [ ] Finance Agent (Yahoo Finance, Bloomberg)
- [ ] WebSocket для real-time обновлений
- [ ] User authentication + персонализация

### 🚀 Future Plans

- [ ] Multi-turn reasoning с длинным контекстом
- [ ] RAG с векторной БД (Qdrant/Weaviate)
- [ ] Fact-checking pipeline с source verification
- [ ] A/B тестирование режимов на production
- [ ] Kubernetes deployment
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Мониторинг и аналитика (Prometheus + Grafana)

## 🤝 Разработка

### Backend (Go)

```bash
cd backend-go

# Установить зависимости
go mod download

# Запустить с hot reload (используйте air или make watch)
go run ./cmd/server/main.go

# Или соберите бинарник
make build
./bin/server
```

### Frontend (Next.js)

```bash
cd frontend

# Dev сервер с hot reload
npm run dev

# Открыть http://localhost:3000
```

### Telegram Bot

```bash
cd backend-go

# Добавить токен в .env
echo "TELEGRAM_BOT_TOKEN=your-token" >> ../../.env

# Запустить бота
go run ./cmd/tgbot/main.go
```

### Структура внесения изменений

1. **Добавление нового агента:**

   - Создать `backend-go/internal/agents/your_agent.go`
   - Имплементировать интерфейс `Agent` с методом `Process()`
   - Добавить в `router.go` логику маршрутизации
   - Обновить `ModeSelector` если нужен новый режим

2. **Добавление нового scraper:**

   - Создать `backend-go/internal/scrapers/your_scraper.go`
   - Имплементировать HTTP-запросы и парсинг HTML
   - Добавить в `SearchClient` в `internal/tools/search_client.go`

3. **Обновление API:**
   - Эндпоинты в `backend-go/internal/api/routes.go`
   - Handlers в `backend-go/internal/api/handlers/`
   - Модели в `backend-go/internal/models/models.go`

## 📄 Лицензия

MIT License

## 👥 Команда

Team Lead - Погодин Иван
Разработчик - Гавриленко Владимир
Бизнес аналитик - Бочарова Станислава

Sber Bootcamp 2025 - Case Study
