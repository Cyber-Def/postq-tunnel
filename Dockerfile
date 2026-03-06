# Stage 1: Build — uses full Go toolchain, produces a statically linked binary
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Cache dependency downloads separately from source compilation
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build two statically linked binaries (no CGO = no libc dependency)
# -s strips symbol table, -w strips debug info → smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/qtunnel ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/qtun ./cmd/qtun/main.go

# Stage 2: Runtime — minimal scratch image, no OS, no shell, no attack surface
FROM scratch

# Copy CA certificates for ACME/Let's Encrypt TLS and outbound HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy both binaries
COPY --from=builder /build/qtunnel /qtunnel
COPY --from=builder /build/qtun /qtun

# Run as non-root (UID 65534 = nobody, standard convention for minimal containers)
USER 65534:65534

# Document exposed ports (actual binding is configured at runtime)
EXPOSE 8080 4443 9090

ENTRYPOINT ["/qtunnel"]
