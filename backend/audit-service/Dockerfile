FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy pre-compiled binary
COPY audit .

# Copy config files
COPY pkg/config ./pkg/config

# Expose service port
EXPOSE 8001

# Run the HTTP server
CMD ["./audit", "serve"]