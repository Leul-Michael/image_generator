# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git and air for hot reloading
RUN apk add --no-cache git && \
    go install github.com/cosmtrek/air@v1.49.0

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with stripped debugging info for smaller binary
RUN go build -ldflags="-s -w" -o main .

# Development stage
FROM golang:1.23-alpine AS development

WORKDIR /app

# Install git and air for hot reloading
RUN apk add --no-cache git && \
    go install github.com/cosmtrek/air@v1.49.0

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Create tmp directory for air
RUN mkdir -p tmp

# Copy the source code
COPY . .

# Expose the application port
EXPOSE 5000

# Run air for hot reloading with polling enabled
CMD ["air", "-c", ".air.toml"]

# Production stage
FROM alpine:3.18 AS production

# Copy the built binary from the builder stage
COPY --from=builder /app/main /main

# Expose the application port
EXPOSE 5000

# Run the application
CMD ["/main"]