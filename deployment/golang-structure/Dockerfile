FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/golangstructure ./cmd/golangstructure
# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/gatewaydummie ./cmd/gatewaydummie
# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/orchestratedummie ./cmd/orchestratedummie

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/ ./bin/

EXPOSE 8080

CMD ["./bin/golangstructure"]
