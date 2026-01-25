# Build stage
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Install migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o bin/server main.go
RUN cp $(go env GOPATH)/bin/migrate bin/migrate

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install ffmpeg and ca-certificates
RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy binaries from builder
COPY --from=builder /app/bin/server /app/bin/server
COPY --from=builder /app/bin/migrate /app/bin/migrate

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Expose port
EXPOSE 8080

# Start command
CMD ["sh", "-c", "./bin/migrate -path migrations -database \"$DATABASE_URL\" up && ./bin/server"]
