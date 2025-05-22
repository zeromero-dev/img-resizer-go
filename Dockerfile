FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install required dependencies for bimg (libvips)
RUN apk add --no-cache \
    build-base \
    vips-dev \
    pkgconfig

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o img-resizer ./cmd/api/

# Build worker 
RUN CGO_ENABLED=1 GOOS=linux go build -o img-resizer-worker ./cmd/worker/

# Create a smaller final image
FROM alpine:3.21

# Install runtime dependencies for libvips
RUN apk add --no-cache \
    vips \
    ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/img-resizer .

# Create storage directory
RUN mkdir -p /app/storage

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./img-resizer"]