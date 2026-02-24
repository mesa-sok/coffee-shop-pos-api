# Build stage
FROM golang:1.26-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for static linking, especially when using Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the .env file if it exists, otherwise it might be mounted or env vars used
# For now, let's assume env vars are passed or .env is mounted.
# But to be safe with the current setup instructions, we can copy .env.example as a fallback or expect .env
# The app handles missing .env gracefully by logging and using defaults/env vars.

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./main"]
