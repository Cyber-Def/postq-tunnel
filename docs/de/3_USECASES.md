# Echte User Stories (Anwendungsfälle)

## 🤖 Anwendungsfall 1: Lokale KI-Automatisierung schützen (n8n, Langflow)
**Problem:** N8n lokal ausführen, aber Webhooks (z.B. von Telegram) empfangen.
**Lösung:** Ein mit Passwort & Whitelist geschützter Tunnel.
```bash
./qtun -sub n8n -local localhost:5678 -allow-ip "198.51.100.0/24" -auth "admin:superHardPass"
```
Zero-Trust-Architektur mit einem Befehl.

## 🧠 Anwendungsfall 2: Sicherer Zugriff auf lokale LLMs (Open WebUI)
**Problem:** Von unterwegs sicher und privat auf Ihr Llama 3 Netz auf dem Mac Studio zugreifen.
**Lösung:** Authentifizierter Tunnel verwehrt Bots den Zugriff.
```bash
./qtun -sub mygpt -local localhost:3000 -auth "aimaster:secureai"
```

## 🛠️ Anwendungsfall 3: "Wormhole" zu einer internen Datenbank
Ein vorübergehender DevOps-Zugang zu Staging-Diensten (z.B. Postgres `5432`), ohne interne Firewalls und NATs kompromittieren zu müssen.

## 📱 Anwendungsfall 4: Schnelles Frontend-Testing (Cross-Device)
Reibungsloses Teilen von lokalem `localhost:5173` für mobile Safari/Chrome Tests über das 4G-Netzwerk. Tunnel stirbt direkt beim Drücken von `Ctrl+C`. Keine Cloud-Reste.
