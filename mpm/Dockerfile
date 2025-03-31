FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./

RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Adjust the path to your main.go file if needed
RUN CGO_ENABLED=0 GOOS=linux go build -o mpm ./cmd/main.go

# Final stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create directory for storing data
RUN mkdir -p /opt/mpm/data

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/mpm .

# Run the application
CMD ["./mpm"]