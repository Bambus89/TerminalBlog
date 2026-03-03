# Changelog

## v1.2.0 – 3. März 2026

### Neue Features

- **Interaktiver TLS-Installer** – Die TLS-Konfiguration im Installer bietet jetzt vier Optionen: kein TLS, automatische Certbot-Erkennung (listet Domains mit Ablaufdatum), ein Terminal-Dateibrowser zum Navigieren und Auswählen von `.pem`/`.crt`/`.key`-Dateien, oder manuelle Pfadeingabe.
- **Graceful TLS-Fallback** – Wenn `cert_file` oder `key_file` auf nicht-existierende Dateien zeigen (z.B. Platzhalter), startet der Server automatisch im HTTP-Modus statt abzustürzen. Eine Warnung wird ins Log geschrieben.
- **Platzhalter in config.json** – Die mitgelieferte `config.json` enthält sprechende Platzhalter (`/PFAD/ZUM/WEBROOT`, `/PFAD/ZUM/ZERTIFIKAT.pem`), die zum Anpassen auffordern und im HTTP-Modus ignoriert werden.

## v1.0.0 – 2. März 2026

### Initiales Release

- Minimaler HTTP/HTTPS-Webserver als einzelnes Go-Binary (nur Stdlib, keine externen Dependencies)
- SPA-Support mit automatischem Fallback auf `index.html`
- Konfiguration über `config.json` (WebRoot, Port, optionale TLS-Zertifikate)
- Sicherheits-Header: X-Content-Type-Options, X-Frame-Options, Referrer-Policy
- Request-Logging mit Methode, Pfad, Statuscode, Dauer und Remote-Adresse
- Verzeichnisauflistung deaktiviert
- Timeouts: Read 15s, Write 30s, Idle 60s
- systemd-Service-Template mit Sicherheitshärtung (NoNewPrivileges, ProtectSystem, PrivateTmp)
- Interaktiver Shell-Installer: fragt Installationspfad, WebRoot, Port, TLS, User/Gruppe, Service-Name ab
- Automatische `CAP_NET_BIND_SERVICE` für Ports < 1024
- Build-Targets: Linux amd64 und arm64
- GPL-3.0 Lizenz
