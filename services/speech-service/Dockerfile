FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies for audio processing
RUN apk add --no-cache git ffmpeg

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o speech-service ./cmd

FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates ffmpeg

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/speech-service .

# Create directories
RUN mkdir -p debug

# Create non-root user
RUN addgroup -g 1001 -S speech && \
    adduser -S -D -H -u 1001 -h /root -s /sbin/nologin -G speech -g speech speech

# Change ownership
RUN chown -R speech:speech /root

USER speech

# Expose port
EXPOSE 50051

# Run the binary
CMD ["./speech-service"]