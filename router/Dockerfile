FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY router/cmd/ ./router/cmd/
COPY router/internal/ ./router/internal/

RUN go build -o /app/bin/router ./router/cmd/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/bin/router /app/router

EXPOSE 3000

ENTRYPOINT ["./router"]
