# PostQ-Tunnel
[![Go CI Build](https://github.com/Cyber-Def/postq-tunnel/actions/workflows/ci.yml/badge.svg)](https://github.com/Cyber-Def/postq-tunnel/actions/workflows/ci.yml)

[Українська](docs/uk/1_INSTALL.md) | [English](docs/en/1_INSTALL.md) | [Deutsch](docs/de/1_INSTALL.md) | [中文](docs/zh/1_INSTALL.md) | [Русский](docs/ru/1_INSTALL.md)

A modern, highly secure, Quantum-Resistant reverse proxy designed to expose your local development and AI agent automation services to the internet.

`PostQ-Tunnel` acts as a drop-in single-binary replacement for services like `ngrok`. It is designed for teams, implementing Identity-Aware layer features, zero-dependency deployments, and post-quantum security.

## Highlights
- **Post-Quantum Cryptography** — The control channel uses `X25519MLKEM768` TLS 1.3 (Go 1.24+), protecting against *Store-Now-Decrypt-Later* quantum decryption attacks.
- **Zero-Dependency** — Pure Go. No Caddy, no OpenSSH, no Python.
- **Yamux Multiplexing** — Millions of HTTP requests over a single PQC-secured stream.
- **Team-Oriented Middleware** — Native IP/CIDR whitelisting, BasicAuth with brute-force lockout, and SSO stubs.
- **Observability** — Built-in Prometheus `/metrics`, `/healthz`, `/readyz` endpoints on port `9090`.
- **Protocol Versioning** — Handshake version check ensures agent ↔ server compatibility.
- **DoS Hardened** — Rate-limited handshakes (10/s per IP), tunnel cap (100), payload size limit (4KB), and subdomain validation.

---

## Architecture

### Connection Flow

```mermaid
sequenceDiagram
    participant User as 👤 End User (Browser)
    participant Edge as 🖥️ Edge Server (qtunnel)
    participant Agent as 🔌 Agent (qtun)
    participant Local as 💻 Local Service

    Agent->>Edge: PQC TLS 1.3 Dial :4443 (X25519MLKEM768)
    Agent->>Edge: Yamux Handshake (version, subdomain, token)
    Edge-->>Agent: Handshake OK → subdomain registered

    User->>Edge: HTTP Request (Host: myapp.domain.com)
    Edge->>Agent: Open Yamux stream for this request
    Agent->>Local: TCP Dial localhost:PORT
    Local-->>Agent: HTTP Response
    Agent-->>Edge: Stream response bytes
    Edge-->>User: HTTP Response
```

### Component Overview

```mermaid
graph TB
    subgraph Internet
        U[👤 User]
    end

    subgraph "Edge Server (qtunnel)"
        EH["HTTP Proxy :8080/443<br/>ReverseProxy + Middleware"]
        EP["PQC Listener :4443<br/>TLS 1.3 / X25519MLKEM768"]
        ER["Registry<br/>subdomain → yamux.Session"]
        EM["Metrics :9090<br/>/healthz /readyz /metrics"]
        EH --> ER
        EP --> ER
    end

    subgraph "Agent Host (qtun)"
        AG["Agent<br/>yamux.Client"]
        LS["Local Service<br/>:3000 / :5432 / any"]
        AG --> LS
    end

    U -->|"HTTP"| EH
    AG -->|"PQC TLS + Yamux"| EP
    EH -.->|"Yamux Stream"| AG
```

### Security Layers

```mermaid
graph LR
    C[Client Connection] --> RL[IP Rate Limit<br/>10 handshakes/s]
    RL --> TLS[TLS 1.3 Handshake<br/>X25519MLKEM768]
    TLS --> PV[Protocol Version<br/>Check v1]
    PV --> SUB[Subdomain Validation<br/>DNS-label regex]
    SUB --> CAP[Tunnel Cap<br/>max 100]
    CAP --> REG[Registry.Register]
    REG --> OK[✅ Tunnel Active]
```

---

## Getting Started

### Build from Source

```bash
# Build the Agent
go build -o qtun ./cmd/qtun/main.go

# Build the Edge Server
go build -o qtunnel ./cmd/server/main.go
```

### Docker Quick Start

```bash
# Clone and run with Docker Compose (local mode, no TLS required)
cd deploy/
docker compose up

# Test the tunnel (agent mounts subdomain "demo")
curl -H "Host: demo.localhost" http://localhost:8080/
```

### Running Edge Server (Production)

```bash
export QTUN_DOMAIN="tunnels.yourdomain.com"
export QTUN_EMAIL="admin@yourdomain.com"
./qtunnel
```

The server binds to:
- `:443` — HTTPS with automatic Let's Encrypt certs (CertMagic)
- `:4443` — PQC TLS agent control plane
- `:9090` — Prometheus metrics + health endpoints

### Running Client Agent

```bash
# Expose local port 3000 as react-preview.tunnels.yourdomain.com
./qtun -server yourvps.com:4443 -sub react-preview -local localhost:3000

# Expose with IP whitelist (only VPN range allowed)
./qtun -server yourvps.com:4443 -sub internal-db -local localhost:5432 -allow-ip 192.168.1.0/24

# Expose with Basic Auth
./qtun -server yourvps.com:4443 -sub staging -local localhost:8080 -auth user:secret
```

---

## Security

See [SECURITY.md](SECURITY.md) for the full threat model, accepted risks, and disclosure policy.

Key protections:
| Layer | Mechanism |
|---|---|
| Transport | TLS 1.3 + ML-KEM-768 (Post-Quantum hybrid) |
| Handshake | Protocol version check, 4KB payload cap |
| Connection | IP rate limit: 10 handshakes/s, max 100 tunnels |
| Auth brute-force | 20 fails / 30s → 1 min lockout (per IP) |
| Subdomain | DNS-label validation (a-z0-9, hyphens, ≤63 chars) |
| Replay / Downgrade | TLS 1.3 forward secrecy + ProtocolVersion field |
