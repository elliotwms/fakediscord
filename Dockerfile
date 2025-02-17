# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download

ARG APP_VERSION="v0.0.0+unknown"
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/elliotwms/fakediscord/internal/fakediscord.Version=${APP_VERSION}" -o /fakediscord ./cmd/main.go

FROM scratch
COPY --from=builder /fakediscord /fakediscord

EXPOSE 8080

ENTRYPOINT ["/fakediscord"]