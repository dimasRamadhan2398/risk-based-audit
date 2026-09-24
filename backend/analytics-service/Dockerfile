FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy pre-compiled binary
COPY analytics .
RUN chmod +x analytics

# Expose service port
EXPOSE 8084

# Run the HTTP server
CMD ["./analytics"]