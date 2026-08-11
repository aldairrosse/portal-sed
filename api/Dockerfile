# Builder stage
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /import ./cmd/import

# Runtime stage
FROM alpine:3.22
RUN adduser -D -u 1000 app
USER app
COPY --from=builder /api /opt/api/api
COPY --from=builder /import /opt/api/import
EXPOSE 8080
ENTRYPOINT ["/opt/api/api"]
