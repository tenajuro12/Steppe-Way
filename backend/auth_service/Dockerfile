FROM golang:1.23-alpine AS builder

WORKDIR /app

# Add wget for wait-for-it script
RUN apk add --no-cache wget

# Download wait-for-it script
RUN wget -O wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh \
    && chmod +x wait-for-it.sh

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth_service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates bash

WORKDIR /root/
COPY --from=builder /app/auth_service .
COPY --from=builder /app/wait-for-it.sh .

ENV PORT=8082
EXPOSE 8082

CMD ["./wait-for-it.sh", "db:5432", "--", "./auth_service"]