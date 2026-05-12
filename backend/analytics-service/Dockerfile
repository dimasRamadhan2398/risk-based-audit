FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o analytics-server cmd/main.go

# Start a new stage from scratch
FROM alpine:latest

WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/analytics-server .

# Expose port
EXPOSE 8084

# Command to run the executable
CMD ["./analytics-server"]
