FROM golang:1.24.3 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Tidy dependencies (ensures go.mod and go.sum are up to date)
RUN go mod tidy

# Build static binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o migrate ./cmd/migrate/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY config.json /root/config.json

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
