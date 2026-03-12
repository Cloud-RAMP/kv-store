FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN go build -o server ./cmd/kv-store/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p internal/store

EXPOSE 3000

CMD ["./server"]
