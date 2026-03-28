FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY store/cmd/ ./store/cmd/
COPY store/internal/ ./store/internal/

RUN go build -o server ./store/cmd/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p /app/data

VOLUME ["/app/data"]

EXPOSE 3000

ENTRYPOINT ["./server"]
