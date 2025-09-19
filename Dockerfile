FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /core .

FROM alpine:latest AS runner

WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates curl

COPY --from=builder /core /app/core

RUN mkdir -p /app/data

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 3280

ENTRYPOINT ["/app/core"]
