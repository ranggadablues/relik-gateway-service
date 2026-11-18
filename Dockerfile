# ==========================
# 1. BUILD STAGE
# ==========================
FROM golang:1.24.7-alpine AS builder

# Install git (needed for go mod download if fetching via VCS)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build statically for target platform (defaults to linux/amd64)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o gateway-service ./cmd

# ==========================
# 2. RUN STAGE
# ==========================
FROM alpine:3.22

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy binary from build stage
COPY --from=builder --chown=app:app /app/gateway-service ./gateway-service

USER app

ENV GATEWAY_SERVICE_PORT=8080

# Expose gateway port
EXPOSE 8080

# Run binary
CMD ["./gateway-service"]
