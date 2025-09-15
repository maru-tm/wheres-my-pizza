FROM golang:1.25 AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o restaurant-system .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/restaurant-system .
COPY config/config.yaml ./config.yaml
COPY migrations ./migrations

EXPOSE 3000 3002

CMD ["./restaurant-system", "--mode=order-service", "--port=3000"]