# site-server

Minimaler, abhängigkeitsfreier Webserver für statische Websites. Ein einzelnes Go-Binary – ideal für minimale Server, Container und VMs.

## Features

- Einzelnes Binary, keine Abhängigkeiten (nur Go-Stdlib)
- SPA-Support (Fallback auf `index.html`)
- Optionales TLS (Certbot / Let's Encrypt)
- Graceful Fallback: Platzhalter-Pfade in der Config → Server startet trotzdem im HTTP-Modus
- Sicherheits-Header (X-Content-Type-Options, X-Frame-Options, Referrer-Policy)
- Request-Logging via stdout / systemd-Journal
- Konfiguration über `config.json`
- systemd-Integration mit interaktivem Installer
- Linux amd64 + arm64 (Raspberry Pi, ARM-Server)

## Schnellstart

```bash
# Kompilieren
make build

# Starten (mit Defaults: Port 8080, WebRoot /var/www/html)
./site-server

# Mit eigener Config
./site-server -config /pfad/zur/config.json

# Version anzeigen
./site-server -v
```

## config.json

Die mitgelieferte `config.json` enthält Platzhalter-Pfade:

```json
{
  "webroot": "/PFAD/ZUM/WEBROOT",
  "port": 8080,
  "cert_file": "/PFAD/ZUM/ZERTIFIKAT.pem",
  "key_file": "/PFAD/ZUM/PRIVKEY.pem"
}
```

**Solange die Platzhalter-Pfade nicht auf echte Dateien zeigen, startet der Server automatisch im HTTP-Modus** und gibt eine Warnung aus. TLS wird erst aktiv, wenn beide Dateien tatsächlich existieren.

### Felder

| Feld | Beschreibung | Beispiel |
|------|-------------|---------|
| `webroot` | Verzeichnis mit den Website-Dateien | `/var/www/html` |
| `port` | TCP-Port | `8080` oder `443` für HTTPS |
| `cert_file` | Pfad zum TLS-Zertifikat (Fullchain) | `/etc/letsencrypt/live/example.com/fullchain.pem` |
| `key_file` | Pfad zum privaten TLS-Schlüssel | `/etc/letsencrypt/live/example.com/privkey.pem` |

Sind `cert_file` und `key_file` leer (`""`) oder zeigen auf nicht vorhandene Dateien, läuft der Server als reiner HTTP-Server.

---

## TLS mit Certbot (Let's Encrypt)

### Schritt 1: Certbot installieren

Certbot ist auf den meisten Distributionen verfügbar:

```bash
# Debian / Ubuntu
sudo apt install certbot

# Fedora / RHEL / CentOS
sudo dnf install certbot

# Arch Linux
sudo pacman -S certbot

# Oder via Snap (distributionsunabhängig)
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
```

### Schritt 2: Zertifikat anfordern

Da der site-server ein eigener Webserver ist, nutzt man den **Standalone-Modus** von Certbot. Dabei startet Certbot kurzzeitig einen eigenen Webserver auf Port 80, um die Domain zu verifizieren.

**Wichtig:** Der site-server muss während der Zertifikatsanforderung gestoppt sein (falls er auf Port 80 oder 443 läuft).

```bash
# site-server stoppen (falls er läuft)
sudo systemctl stop site-server

# Zertifikat anfordern
sudo certbot certonly --standalone -d example.com -d www.example.com

# site-server wieder starten
sudo systemctl start site-server
```

Certbot legt die Zertifikate unter `/etc/letsencrypt/live/` ab:

```
/etc/letsencrypt/live/example.com/
├── fullchain.pem    ← Das ist cert_file (Zertifikat + Zwischenzertifikate)
├── privkey.pem      ← Das ist key_file  (Privater Schlüssel)
├── cert.pem         ← (Nur das Zertifikat selbst – nicht verwenden)
└── chain.pem        ← (Nur die Zwischenzertifikate – nicht verwenden)
```

**Wichtig:** Immer `fullchain.pem` verwenden, nicht `cert.pem`. Der Fullchain enthält sowohl das Zertifikat als auch die Zwischenzertifikate, die Browser zur Validierung brauchen.

### Schritt 3: config.json anpassen

Ersetze die Platzhalter durch die tatsächlichen Pfade:

```json
{
  "webroot": "/var/www/html",
  "port": 443,
  "cert_file": "/etc/letsencrypt/live/example.com/fullchain.pem",
  "key_file": "/etc/letsencrypt/live/example.com/privkey.pem"
}
```

Ersetze `example.com` durch deine eigene Domain.

### Schritt 4: Dateiberechtigungen

Die Let's-Encrypt-Zertifikate gehören `root`. Der site-server braucht Leserechte:

```bash
# Option A: Den Service-User zur Gruppe ssl-cert hinzufügen (Debian/Ubuntu)
sudo usermod -aG ssl-cert dein-user

# Option B: Leserechte für die Live-Verzeichnisse setzen
sudo chmod 750 /etc/letsencrypt/live/
sudo chmod 750 /etc/letsencrypt/archive/
sudo chgrp -R ssl-cert /etc/letsencrypt/live/ /etc/letsencrypt/archive/
```

Falls der Service als eigener User läuft (z.B. `www-data`):

```bash
sudo usermod -aG ssl-cert www-data
sudo systemctl restart site-server
```

### Schritt 5: Automatische Erneuerung

Certbot erneuert Zertifikate automatisch per Cron/Timer. Damit der site-server das neue Zertifikat lädt, muss er nach der Erneuerung neugestartet werden:

```bash
# Certbot-Hook einrichten
sudo tee /etc/letsencrypt/renewal-hooks/post/restart-site-server.sh > /dev/null <<'EOF'
#!/bin/bash
systemctl restart site-server
EOF
sudo chmod +x /etc/letsencrypt/renewal-hooks/post/restart-site-server.sh
```

Erneuerung testen:

```bash
sudo certbot renew --dry-run
```

---

## Betrieb ohne TLS

Für lokale Nutzung, hinter einem Reverse Proxy (nginx, Caddy, Traefik) oder im Heimnetz genügt HTTP:

```json
{
  "webroot": "/home/user/website",
  "port": 3000
}
```

Die Zertifikats-Felder können leer bleiben oder ganz weggelassen werden.

---

## Installation (Linux + systemd)

```bash
# Binary bauen
make linux

# Installer starten
sudo ./install.sh
```

Der Installer fragt interaktiv nach:

1. **Installationsverzeichnis** – wohin Binary + Config kopiert werden (Default: `/opt/site-server`)
2. **WebRoot** – Verzeichnis mit `index.html` und den JSON-Dateien
3. **Port** – TCP-Port für den Server
4. **TLS-Konfiguration** – Auswahl aus vier Optionen:
   - Kein TLS (reiner HTTP-Server)
   - Let's Encrypt / Certbot automatisch erkennen (listet vorhandene Domains)
   - Dateibrowser zum Navigieren und Auswählen der Zertifikatsdateien
   - Pfade direkt eingeben
5. **Service-Benutzer** – unter welchem User der Server läuft
6. **Service-Name** – Name des systemd-Service (Default: `site-server`)

Nach der Installation:

```bash
sudo systemctl start site-server      # Starten
sudo systemctl stop site-server       # Stoppen
sudo systemctl restart site-server    # Neustarten
sudo systemctl status site-server     # Status
sudo journalctl -u site-server -f     # Logs live verfolgen
```

## Build-Targets

```bash
make build        # Aktuelle Plattform
make linux        # Linux amd64
make linux-arm64  # Linux arm64 (Raspberry Pi, ARM-Server)
make all          # Beide Linux-Architekturen
make clean        # Aufräumen
```

## Releases

| Version | Datum | Highlights |
|---------|-------|------------|
| v1.2 | 3. März 2026 | Interaktiver TLS-Installer (4 Optionen: kein TLS, Certbot-Erkennung, Dateibrowser, manuelle Eingabe), Graceful TLS-Fallback bei fehlenden Zertifikaten, Platzhalter in config.json |
| v1.0 | 2. März 2026 | Initiales Release – HTTP/HTTPS-Server, SPA-Fallback, Sicherheits-Header, Request-Logging, systemd-Service, interaktiver Installer |

Siehe [CHANGELOG.md](CHANGELOG.md) für Details.

## Projektstruktur

```
site-server/
├── main.go                    # Webserver (~150 Zeilen, nur Go-Stdlib)
├── config.json                # Konfiguration (mit Platzhaltern)
├── install.sh                 # Interaktiver Installer mit Dateibrowser
├── site-server.service.tpl    # systemd-Service-Template
├── Makefile                   # Build-Targets
├── go.mod                     # Go-Modul (keine Dependencies)
├── CHANGELOG.md               # Versionshistorie
├── README.md                  # Diese Datei
└── LICENSE                    # GPL-3.0
```

## Lizenz

GPL-3.0 – Siehe [LICENSE](LICENSE)
