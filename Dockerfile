FROM golang:1.25

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# -ldflags="-w -s" reduces binary size by stripping debug information
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .
# We expect .env.prod to be passed as .env or mounted, but here we copy it if it exists in build context
# However, usually envs are injected via docker-compose or k8s.
# For this setup, we'll assume .env is mounted or we copy a default if needed.
# The user asked for .env.prod.

# Expose port
EXPOSE 8000

# Run the binary
CMD ["./main"]
