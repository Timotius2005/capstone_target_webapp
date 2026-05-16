FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static binary — no CGO so the runtime image needs no libc.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/main.go

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/.env.example .env.example

USER appuser

EXPOSE 8080

CMD ["./server"]
