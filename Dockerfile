# Build stage
CMD ["./main"]
# Run the application

EXPOSE 8080
# Expose port

COPY --from=builder /app/config.yaml .
COPY --from=builder /app/main .
# Copy binary from builder

WORKDIR /root/

RUN apk --no-cache add ca-certificates

FROM alpine:latest
# Runtime stage

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api
# Build the application

COPY . .
# Copy source code

RUN go mod download
COPY go.mod go.sum ./
# Copy go mod files

WORKDIR /app

FROM golang:1.25-alpine AS builder

