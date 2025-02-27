# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install runtime dependencies including pdftotext (poppler-utils)
RUN apk add --no-cache \
    poppler-utils \
    poppler \
    poppler-dev \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    tesseract-ocr-data-vie \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy any additional required files
COPY .env .

# Create upload directory
RUN mkdir -p /app/upload

# Expose port
EXPOSE 8888

# Set environment variables
ENV GIN_MODE=release

# Verify pdftotext installation
RUN pdftotext -v

# Run the application
CMD ["./main", "start"]