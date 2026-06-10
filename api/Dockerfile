# Builder stage
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/server

# Runtime stage
FROM alpine:3.19
RUN adduser -D -u 1000 app
USER app
COPY --from=builder /api /opt/api/api
EXPOSE 8080
ENTRYPOINT ["/opt/api/api"]
