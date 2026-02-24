# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
# Copy .env.example as .env (user should override this in production)
COPY .env.example .env

# Expose port 8080
EXPOSE 8080

# Command to run
CMD ["./main"]
