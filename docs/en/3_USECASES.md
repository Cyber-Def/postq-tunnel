# Real User Stories (Use Cases)

*How engineers and AI architects solve daily problems with PostQ-Tunnel.*

## 🤖 Use Case 1: Protecting Local AI Automation (n8n, Langflow)
**Context (Problem):** You host complex automated workflows via [n8n](https://n8n.io/) or Langflow on a local Mac Studio. You absolutely need external webhooks (e.g., Stripe payments or Telegram Bot API). But exposing your raw n8n dashboard to the public risks RCE attacks on your Mac.
**Solution:**
Secure the tunnel with the built-in IP whitelist and BasicAuth.
```bash
# Allow UI access only from the corporate VPN, protected by a password:
./qtun -server yourvps.com:4443 -sub n8n -local localhost:5678 -allow-ip "198.51.100.0/24" -auth "admin:superHardPass"
```
**Result:** The admin panel is completely hidden behind the Edge server. Standard Zero-Trust architecture is achieved in a single command.

## 🧠 Use Case 2: Secure Access to Personal LLMs (Open WebUI / OpenClaw)
**Context (Problem):** You're running Llama 3 locally via Ollama + Open WebUI. You want to query your chatbot away from home on your iPhone, but you don't want someone scraping your URL and mining queries on your GPU.
**Solution:**
Launch an authenticated tunnel.
```bash
./qtun -sub mygpt -local localhost:3000 -auth "aimaster:secureai"
```
**Result:** Accessing `https://mygpt.tunnels...` from your phone prompts for a password block directly at the proxy layer. ISPs only observe quantum-encrypted TLS noise.

## 🛠️ Use Case 3: "Wormhole" to an Internal Database
**Context (Problem):** A staging PostgreSQL database is behind strict corporate NAT with no static IP. A contractor (DevOps) needs temporary access from home.
**Solution:**
Run qtun from inside the corporate network to your VPS:
```bash
./qtun -sub testdb -local localhost:5432 -auth "devops:token"
```
**Result:** The remote DevOps can connect securely (e.g. via DataGrip/DBeaver). Internal firewalls are never breached.

## 📱 Use Case 4: Rapid Cross-Device Preview
**Context (Problem):** You're building a React/Vite frontend on `localhost:5173`. You need to preview animations on an iPad over a 4G connection.
**Solution:**
```bash
./qtun -sub preview -local localhost:5173
```
**Result:** The second you finish testing, you hit `Ctrl+C`. The tunnel vanishes instantly with Graceful Shutdown. No Vercel deployment junk is left behind.
