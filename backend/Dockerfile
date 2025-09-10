# Stage 1: Dependencies
FROM golang:1.25-alpine AS deps
WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies (cached layer)
RUN go mod download

# Stage 2: Builder
FROM golang:1.25-alpine AS builder
WORKDIR /build

# Copy dependencies from previous stage
COPY --from=deps /go/pkg /go/pkg

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH

# Build the binary
# CGO_ENABLED=0 for static binary
# -ldflags="-s -w" to strip debug info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.BuildTime=${BUILD_TIME} \
        -X main.CommitHash=${COMMIT_HASH}" \
    -o backend \
    ./cmd/main.go

# Stage 3: Minimal runtime
FROM alpine:latest AS export
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Create necessary directories
RUN mkdir -p /app/data && \
    chown -R appuser:appuser /app

USER appuser
WORKDIR /app

# Copy the binary
COPY --from=builder /build/backend /app/backend

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the backend
ENTRYPOINT ["/app/backend"]
