# PostQ-Tunnel

[🇺🇦 Українська](docs/uk/1_INSTALL.md) | [🇬🇧 English](docs/en/1_INSTALL.md) | [🇩🇪 Deutsch](docs/de/1_INSTALL.md) | [🇨🇳 中文](docs/zh/1_INSTALL.md) | [🇷🇺 Русский](docs/ru/1_INSTALL.md) 

A modern, highly secure, Quantum-Resistant reverse proxy designed to expose your local development and AI agents automation services to the internet. 

`PostQ-Tunnel` acts as a drop-in single-binary replacement for services like `ngrok`. It is designed for teams, implementing Identity-Aware layer features, zero-dependency deployments, and post-quantum security.

## Highlights
- **Post-Quantum Cryptography** The underlying control channel uses `X25519MLKEM768` TLS 1.3 (Go 1.24+ standard), protecting the tunneled payload today against *Store-Now-Decrypt-Later* quantum decryption attacks.
- **Zero-Dependency** Drops the Caddy, OpenSSH, and Python dependencies. Server and Client are combined in a pure Go codebase.
- **Yamux Multiplexing** Routing millions of requests over a single constant stream, vastly reducing connection latency.
- **Team-Oriented Middleware** Native IP/CIDR VPN Whitelisting, BasicAuth, and placeholders for SSO directly in the Edge Proxy (`internal/middleware`).
- **Observability** Built-in Prometheus `/metrics` handler exposing `qtun_active_tunnels`.

## Getting Started

Because it’s standard Go code, you build it via:

```bash
# Build the Agent (qtun)
go build -o qtun ./cmd/qtun/main.go

# Build the Edge Server
go build -o qtunnel ./cmd/server/main.go
```

### Running Edge Server

```bash
# On your Linux VPS / Server
export QTUN_DOMAIN="tunnels.yourdomain.com"
export QTUN_EMAIL="admin@yourdomain.com"
./qtunnel
```
The server will bind to `443` holding CertMagic domains, `4443` for accepting PQC client agents, and `9090` for prometheus metrics.

### Running Client Agent

```bash
# Expose your local port 3000 to the domain react-preview.tunnels.yourdomain.com
./qtun -server yourvps.com:4443 -sub react-preview -local localhost:3000

# Expose your local DB requiring the incoming requester to be behind the 192.168.1.0/24 VPN
./qtun -server yourvps.com:4443 -sub local-db -local localhost:5432 -allow-ip 192.168.1.0/24
```
