# Use Golang image for building
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the service executable
RUN go build -o profile_service ./cmd/main.go

# Create lightweight production image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/profile_service .

# Expose the port for the service
ENV PORT=8084
EXPOSE 8084

# Run the service
CMD ["./profile_service"]
