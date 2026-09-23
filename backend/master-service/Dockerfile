FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy pre-compiled binary
COPY master .

# Copy config files
COPY pkg/config ./pkg/config

# Expose port
EXPOSE 8002

# Run the application
CMD ["./master", "serve"]
