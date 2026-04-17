FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=1
RUN go run github.com/a-h/templ/cmd/templ@v0.3.1001 generate
RUN APP_ENV=production go run ./cmd/assets
RUN go build -o server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./server"]
