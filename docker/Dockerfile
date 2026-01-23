FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o /codeworkers ./cmd/worker

FROM alpine:3.19

RUN apk add --no-cache docker-cli

COPY --from=builder /codeworkers /codeworkers

EXPOSE 8080

CMD ["/codeworkers"]
