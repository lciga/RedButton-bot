FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /build/redbutton-bot ./cmd/bot

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S redbutton \
    && adduser -S -G redbutton redbutton

WORKDIR /app

COPY --from=builder /build/redbutton-bot /app/redbutton-bot
COPY --chown=redbutton:redbutton tasks /app/tasks

USER redbutton

STOPSIGNAL SIGTERM

ENTRYPOINT ["/app/redbutton-bot"]
