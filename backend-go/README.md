# Research Pro Mode - Go Backend

High-performance Go backend for the Research Pro Mode multi-agent search assistant.

## 🚀 Features

- **Fast & Efficient**: Go's concurrency and performance
- **Simple Mode**: Quick search with minimal overhead
- **Pro Mode**: Deep analysis with context awareness
- **Chat Support**: Conversation history and context
- **Mode Selector**: Automatic mode detection
- **REST API**: Clean JSON API with Gin framework
- **Database**: GORM with SQLite/PostgreSQL support
- **LLM Integration**: OpenAI/Qwen compatible

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL or SQLite
- OpenAI API key or Qwen endpoint

## 🛠️ Installation

### 1. Clone and Setup

```bash
cd backend-go

# Download dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env with your API keys
nano .env
```

### 2. Run

```bash
# Development mode
make run

# Or directly
go run cmd/server/main.go

# Build binary
make build
./bin/server
```

### 3. Docker (Optional)

```bash
# Build image
make docker-build

# Run container
make docker-run
```

## 📁 Project Structure

```
backend-go/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── agents/
│   │   ├── router.go         # Route between Simple/Pro
│   │   ├── mode_selector.go  # Auto mode selection
│   │   ├── simple_agent.go   # Simple mode logic
│   │   └── pro_agent.go      # Pro mode logic
│   ├── api/
│   │   ├── routes.go         # Route setup
│   │   └── handlers/
│   │       ├── health.go     # Health check
│   │       ├── search.go     # Search endpoint
│   │       └── chat.go       # Chat endpoints
│   ├── config/
│   │   └── config.go         # Configuration
│   ├── database/
│   │   └── database.go       # Database models & setup
│   ├── models/
│   │   └── models.go         # Request/Response models
│   └── tools/
│       ├── search_client.go  # Search API client
│       └── llm_client.go     # LLM client
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```

## 🔧 API Endpoints

### Health Check

```bash
GET /api/health
```

### Search

```bash
POST /api/search
Content-Type: application/json

{
  "query": "What is quantum computing?",
  "mode": "auto"  # auto, simple, or pro
}
```

### Chat - Create Session

```bash
POST /api/chat/session
Content-Type: application/json

{
  "mode": "pro"
}
```

### Chat - Send Message

```bash
POST /api/chat/session/:session_id/message
Content-Type: application/json

{
  "query": "Tell me more about that",
  "mode": "pro"
}
```

### Chat - Get History

```bash
GET /api/chat/session/:session_id
```

### Chat - Delete Session

```bash
DELETE /api/chat/session/:session_id
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/agents/...

# Run with coverage
go test -cover ./...
```

## 📊 Go Library Equivalents

| Python Package | Go Equivalent | Purpose         |
| -------------- | ------------- | --------------- |
| FastAPI        | Gin           | Web framework   |
| Pydantic       | validator     | Data validation |
| SQLAlchemy     | GORM          | ORM             |
| httpx          | resty         | HTTP client     |
| redis          | go-redis      | Redis client    |
| langchain      | langchaingo   | LLM framework   |
| beautifulsoup4 | goquery       | HTML parsing    |
| requests       | net/http      | HTTP requests   |

## 🔥 Performance Benefits

- **10x faster startup** compared to Python
- **Lower memory footprint** (~20MB vs ~100MB)
- **Better concurrency** with goroutines
- **Compiled binary** - no runtime dependencies
- **Native performance** for CPU-intensive tasks

## 🚀 Deployment

### Binary

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/server-linux cmd/server/main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o bin/server-mac cmd/server/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o bin/server.exe cmd/server/main.go
```

### Docker

```bash
docker build -t research-pro-backend .
docker run -p 8000:8000 --env-file .env research-pro-backend
```

### Docker Compose

```yaml
backend-go:
  build: ./backend-go
  ports:
    - "8000:8000"
  environment:
    - DATABASE_URL=postgres://user:pass@db:5432/research_pro
    - OPENAI_API_KEY=${OPENAI_API_KEY}
  depends_on:
    - db
    - tavily-adapter
```

## 🛠️ Development

### Hot Reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
make lint
```

## 📝 Environment Variables

See `.env.example` for all available configuration options.

Key variables:

- `PORT` - Server port (default: 8000)
- `DATABASE_URL` - Database connection string
- `OPENAI_API_KEY` - OpenAI API key
- `TAVILY_URL` - Search service URL

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## 📄 License

MIT License

## 👥 Team

Sber Bootcamp 2025 - Case Study
