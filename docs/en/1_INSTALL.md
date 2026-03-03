# Setup and Installation Guide (PostQ-Tunnel)

## 0. Infrastructure Requirements
- Installed **Go 1.24** or higher (Go 1.26 is recommended).
- **Public VPS Server** with a dedicated public IP address (e.g., standard Ubuntu/Debian VM).
- **Domain Name** (e.g., `yourdomain.com`) pointing via A-record to your VPS IP.
  - *Important:* A **Wildcard record** `*.yourdomain.com` (or CNAME) is required for tunnels to dynamically claim subdomains (e.g., `api.yourdomain.com`).

## 1. Compiling Binaries
The project is built on the *Zero-Dependency* paradigm, so you only need to compile two lightweight binaries. In the project root directory, run:

```bash
# Build the public Edge Server (deployed to the VPS)
go build -o qtunnel ./cmd/server/main.go

# Build the local client agent (stays on your PC/Mac)
go build -o qtun ./cmd/qtun/main.go
```

## 2. VPS Configuration (Public Gateway)
Transfer the compiled `qtunnel` file to your server. This server acts as a relay proxy and automatically issues valid SSL certificates.

Set variables and start the process (using `systemd` or `tmux` is recommended):
```bash
export QTUN_DOMAIN="tunnels.yourdomain.com"
export QTUN_EMAIL="admin@yourdomain.com" # Required for Let's Encrypt certificates

# Note: Binding public ports (80 and 443) usually requires root privileges.
sudo -E ./qtunnel
```
**Ports that MUST be open in the VPS firewall (UFW/Iptables):**
- `80` and `443` — Public HTTP/HTTPS (for users/browsers).
- `4443` — PQC TLS Transport (exclusively for your qtun agents' secure connections).
- `9090` — (Optional) Prometheus metrics tracking (`/metrics`).

## 3. Running the Tunnel Locally (Your PC/Mac)
On your laptop (or smart home hub) where your services are located:

**A) Instant Port Forwarding (like ngrok)**
Showcase a local React server:
```bash
./qtun -server YOUR_VPS_IP:4443 -sub react -local localhost:3000
```
*(Result: accessible via `https://react.tunnels.yourdomain.com`)*

**B) Forwarding with Basic Auth Protection**
```bash
./qtun -server YOUR_VPS_IP:4443 -sub internal-api -local localhost:8080 -auth "admin:supersecret"
```

**C) Forwarding with IP Filtering (White-List)**
Restrict access only to corporate VPN devices:
```bash
./qtun -server YOUR_VPS_IP:4443 -sub admin-db -local localhost:5432 -allow-ip "192.168.1.0/24"
```
