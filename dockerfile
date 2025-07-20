FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o myapp 

# ---- Final image ----
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/myapp .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8080

CMD ["./myapp"]
