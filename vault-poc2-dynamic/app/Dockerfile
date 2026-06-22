# =========================================================
# Stage 1: Build the Go binary
# =========================================================
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY main.go .
# Initialize a temporary go module and compile a static binary
RUN go mod init poc2 && go build -o main main.go

# =========================================================
# Stage 2: Create the final, ultra-minimal runner image
# =========================================================
FROM alpine:latest
WORKDIR /app

# Copy ONLY the compiled binary. No source code, no build tools.
COPY --from=builder /app/main .

# Create the exact directory where Kubernetes will mount the Vault Secret
RUN mkdir -p /app/secrets

EXPOSE 8080

# Run the binary directly. 
# If /app/secrets/firebase-sa.json is missing, the container will instantly crash.
CMD ["./main"]