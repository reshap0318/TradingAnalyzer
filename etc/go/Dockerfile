# Dockerfile for TradingBot Go Application with Air (Hot Reload)
FROM golang:1.25.0-alpine

# Install air for hot reload and other dependencies
RUN apk add --no-cache \
    git \
    curl \
    gcc \
    musl-dev \
    tzdata \
    && go install github.com/air-verse/air@latest

# Set timezone to Asia/Jakarta (WIB)
ENV TZ=Asia/Jakarta

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY be/go.mod be/go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY be/ ./

# Copy air configuration
COPY be/.air-linux.toml .air.toml

# Expose application port
EXPOSE 9000

# Set environment variable to use testnet by default
ENV BINANCE_IS_TESTNET=true
ENV PORT=9000

# Run air for hot reload
CMD ["go", "run", "cmd/main/main.go"]
