# Use Golang image for building
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

RUN go build -o accommodation_service ./cmd/main.go

# Create lightweight production image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/accommodation_service .

ENV PORT=8089
EXPOSE 8089
# Run the service
CMD ["./accommodation_service"]
