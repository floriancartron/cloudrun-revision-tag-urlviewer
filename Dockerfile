# Build Stage
FROM golang:1.23 AS builder

WORKDIR /app

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(go env GOARCH) go build -o cloudrun-revision-tag-urlviewer .

FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/floriancartron/cloudrun-revision-tag-urlviewer
# Use a non-root user for security
USER nobody

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder --chown=nobody:nobody /app/cloudrun-revision-tag-urlviewer .
COPY ./index.html .
EXPOSE 8080

# Set the entry point to the binary
CMD ["./cloudrun-revision-tag-urlviewer"]

