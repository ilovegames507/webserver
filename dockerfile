# Build stage
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 go build -tags netgo -ldflags '-s -w' -o chatserver

# Final minimal image
FROM alpine:latest

WORKDIR /app

# Copy the statically built binary
COPY --from=builder /app/chatserver .

# Expose the port your app listens on
EXPOSE 8000

# Run the binary
CMD ["./chatserver"]
