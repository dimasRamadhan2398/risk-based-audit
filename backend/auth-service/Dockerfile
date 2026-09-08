FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy pre-compiled binary
COPY auth .

# Copy docker config as runtime config
COPY pkg/config/config.docker.yaml ./pkg/config/config.yaml

# Expose service port
EXPOSE 8001

# Run the HTTP server
CMD ["./auth", "--config", "./pkg/config/config.yaml", "serve"]