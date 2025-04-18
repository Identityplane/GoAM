# Build stage
FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Install swag tool
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag init --generalInfo internal/web/admin_api/swagger_info.go --dir ./ --output ./internal/web/swagger-ui
RUN CGO_ENABLED=0 GOOS=linux go build -o goiam ./cmd

# Final image
FROM alpine

WORKDIR /app

COPY --from=builder /app/goiam .
COPY --from=builder /app/cmd/goiam.db .
COPY --from=builder /app/config ./config

ENV GOIAM_CONFIG_PATH=/app/config
CMD ["./goiam"]