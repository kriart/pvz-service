# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o pvz-api ./cmd/pvz-api && \
    go build -o pvz-grpc ./cmd/pvz-grpc && \
    go build -o pvz-migrator ./cmd/pvz-migrator

# Run stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/pvz-api /app/pvz-api
COPY --from=builder /app/pvz-grpc /app/pvz-grpc
COPY --from=builder /app/pvz-migrator /app/pvz-migrator
COPY ./migrations /app/migrations
EXPOSE 8080 3000 9000
CMD ["/app/pvz-api"]
