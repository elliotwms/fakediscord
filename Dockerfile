# syntax=docker/dockerfile:1
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /fakediscord ./cmd/main.go

FROM scratch
COPY --from=builder /fakediscord /fakediscord

EXPOSE 8080

ENTRYPOINT ["/fakediscord"]