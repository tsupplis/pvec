# Build stage
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pvec .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/pvec .

# Run as non-root user
RUN adduser -D -u 1000 pvec && \
    chown -R pvec:pvec /app
USER pvec

ENTRYPOINT ["/app/pvec"]
