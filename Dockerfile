# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies for CGO (required by SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with build date injected
RUN CGO_ENABLED=1 go build \
    -ldflags "-X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o card-shoggoths ./cmd/card-shoggoths-server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and sqlite dependencies
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/card-shoggoths .

# Copy static assets
COPY --from=builder /build/static ./static

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/card-shoggoths"]
