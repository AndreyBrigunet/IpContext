# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Download Go modules
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Ensure go.mod and go.sum reflect actual imports
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /ipapi ./cmd/ipapi

# Runtime stage
FROM alpine:latest AS runner

WORKDIR /app

# Install timezone data
RUN apk --no-cache add tzdata ca-certificates curl

# Copy the binary from builder
COPY --from=builder /ipapi /app/ipapi

# Create directory for GeoIP databases
RUN mkdir -p /app/data

# Run as non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose the port the app runs on
EXPOSE 3280

# Command to run the application
ENTRYPOINT ["/app/ipapi", "--db-path", "/app/data"]
