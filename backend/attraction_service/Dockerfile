# Use Golang image for building
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

RUN go build -o attraction_service ./cmd/main.go

# Create lightweight production image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/attraction_service .

ENV PORT=8085
EXPOSE 8085
# Run the service
CMD ["./attraction_service"]
