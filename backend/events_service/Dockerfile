# Use Golang image for building
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the service executable
RUN go build -o events_service ./cmd/main.go

# Create lightweight production image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/events_service .

# Expose the port for the service
ENV PORT=8083
EXPOSE 8083

# Run the service
CMD ["./events_service"]
