FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /core .

FROM alpine:latest AS runner

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add tzdata ca-certificates curl && \
    rm -rf /var/cache/apk/*

COPY --from=builder /core /app/core

# Create data directory and non-root user
RUN mkdir -p /app/data && \
    adduser -D -g '' appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 3280

ENTRYPOINT ["/app/core"]
