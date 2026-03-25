FROM golang:alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=1
RUN APP_ENV=production go run ./cmd/assets/main.go
RUN go build -o server ./cmd/server
RUN go build -o seed ./cmd/seed

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/seed .
COPY --from=builder /app/static ./static
COPY --from=builder /app/models ./models
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
COPY --from=builder /app/internal/db/seeds ./internal/db/seeds

EXPOSE 8080

CMD ["./server"]
