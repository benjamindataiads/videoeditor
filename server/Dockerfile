FROM golang:1.22-alpine AS builder

# Install build dependencies and FFmpeg in one layer
RUN apk add --no-cache --update \
    ffmpeg \
    ffmpeg-dev \
    ca-certificates \
    git \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage - smaller runtime image
FROM alpine:latest

# Install only runtime dependencies
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create necessary directories
RUN mkdir -p data/uploads data/exports data/work

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
