# Installations- und Einrichtungsanleitung (PostQ-Tunnel)

## 0. Infrastrukturanforderungen
- Installiertes **Go 1.24** oder höher (Go 1.26 wird empfohlen).
- **Öffentlicher VPS-Server** mit einer zugewiesenen öffentlichen IP-Adresse (z. B. eine Standard-Ubuntu/Debian-VM).
- **Domainname** (z.B. `yourdomain.com`), der per A-Record auf die VPS-IP verweist.
  - *Wichtig:* Ein **Wildcard-Eintrag** `*.yourdomain.com` (oder CNAME) ist erforderlich, damit Tunnel dynamisch Subdomains (z.B. `api.yourdomain.com`) verwenden können.

## 1. Binärdateien kompilieren
Das Projekt basiert auf dem *Zero-Dependency* Prinzip. Sie müssen nur zwei leichtgewichtige Binärdateien kompilieren:

```bash
# Kompilieren des öffentlichen Edge-Servers (für den VPS)
go build -o qtunnel ./cmd/server/main.go

# Kompilieren des lokalen Client-Agenten (bleibt auf Ihrem PC/Mac)
go build -o qtun ./cmd/qtun/main.go
```

## 2. VPS-Konfiguration (Öffentliches Gateway)
Exportieren Sie Variablen und starten Sie den Prozess (mit `systemd` oder `tmux`):
```bash
export QTUN_DOMAIN="tunnels.yourdomain.com"
export QTUN_EMAIL="admin@yourdomain.com" 

# Achtung: Das Binden der öffentlichen Ports (80 und 443) erfordert Root-Rechte.
sudo -E ./qtunnel
```
**Ports, die auf dem VPS geöffnet sein MÜSSEN:**
- `80` und `443` — Öffentliches HTTP/HTTPS.
- `4443` — PQC TLS Transport (Nur für die sicheren Verbindungen Ihrer qtun-Agenten).
- `9090` — (Optional) Prometheus-Metriken.

## 3. Tunnel lokal ausführen (Ihr PC/Mac)
Beispiel für eine lokale React-Anwendung:
```bash
./qtun -server YOUR_VPS_IP:4443 -sub react -local localhost:3000
```

Mit sicherem Passwort-Schutz:
```bash
./qtun -server YOUR_VPS_IP:4443 -sub my-dashboard -local localhost:8080 -auth "admin:supersecret"
```
