# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24.2 AS builder
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

# Explicitly declare ARG again to be safe
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o goiam ./cmd

# Final image
FROM alpine

WORKDIR /app

COPY --from=builder /app/goiam .
COPY --from=builder /app/config ./config

ENV GOIAM_CONFIG_PATH=/app/config
CMD ["./goiam"]