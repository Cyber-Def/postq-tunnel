# Coole Features und Vorteile (PostQ-Tunnel)

## 🛡️ Beispiellose Sicherheitsschicht
1. **Hybride Post-Quanten-Kryptographie (X25519MLKEM768):**
Der Datenverkehr ist vor **"Store Now, Decrypt Later"**-Angriffen geschützt.
2. **Kein Risiko für das Server-OS:**
Anders als bei OpenSSH-Tunneln gibt es keine Linux-Benutzer auf dem Server. Alles wird über Multiplexing auf Applikationsebene abgewickelt.

## 🚀 Beeindruckende Leistung
1. **Yamux-Multiplexing (Wie HTTP/2):**
Ein einziger beständiger Tunnel wird geöffnet, was eine enorme Einsparung an Arbeitsspeicher bewirkt und Latenzen beseitigt.
2. **Blitzschnelles In-Memory-Routing:**
Der Proxyserver verwaltet aktive Tunnel im RAM.
3. **Absolute Zero-Dependency:**
Kein Docker, kein Python. Alles basiert auf statischen, portablen Binärdateien.

## 👥 Entwickelt für Teams
1. **IP- & CIDR-Whitelisting:** Zugriff nur über das Firmen-VPN erlauben (erfolgt L7).
2. **Dynamic Basic Auth:** Schützen Sie lokale Dashboards im Handumdrehen per CLI Flag (`--auth "user:pass"`).
3. **SSO Ready:** Architektur vorbereitet für Identity-Aware Proxys wie GitHub Login.
4. **Resilienz:** Automatischer "Exponential Backoff Reconnect" bei Internetabbrüchen.
