# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Enable CGO if needed, but alpine is usually fine without it
ENV CGO_ENABLED=0

# Copy the entire source code first (needed for local module replacements)
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o quizserver cmd/quizserver/main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /root/

# Copy the binary and config from the builder stage
COPY --from=builder /app/quizserver .
COPY --from=builder /app/config.yml .

# Expose ports based on config.yml
# HTTP: 8082, WS: 8083, TCP: 8080
EXPOSE 8080 8082 8083

# Run the application
CMD ["./quizserver"]
