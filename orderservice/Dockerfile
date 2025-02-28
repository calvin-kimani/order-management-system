FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache netcat-openbsd && \
    chmod +x wait-for.sh && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o orderservice .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates netcat-openbsd
COPY --from=builder /app/orderservice .
COPY --from=builder /app/wait-for.sh .
EXPOSE 8080
CMD ["./wait-for.sh", "postgres:5432", "--", "./orderservice"]