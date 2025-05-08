FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o fav_service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/fav_service .

ENV PORT=8088
EXPOSE 8088

CMD ["./fav_service"]
