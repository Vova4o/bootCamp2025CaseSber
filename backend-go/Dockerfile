FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk update && apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:3.19

RUN apk update && apk add --no-cache ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env.example

EXPOSE 8000

CMD ["./main"]
