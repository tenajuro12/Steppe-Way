# Use Golang image for building
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

RUN go build -o food_service ./cmd/main.go

# Create lightweight production image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/food_service .

ENV PORT=8090
EXPOSE 8090
# Run the service
CMD ["./food_service"]
