# Personal Terminal Website

Eine moderne, responsive Einzelseiten-Website im Terminal-Stil mit Gruvbox-Theme – komplett mit TUI-Verwaltungstool und eigenem Webserver. Keine externen Abhängigkeiten, keine Datenbanken, keine Frameworks. Drei Komponenten, ein Ökosystem.

## Übersicht

| Komponente | Beschreibung | Sprache | Version |
|------------|-------------|---------|---------|
| **Website** (`index.html`) | Single-File SPA im Terminal-Design | HTML/CSS/JS | v4.0 |
| **Site Manager** (`site-manager/`) | TUI zur Verwaltung aller Inhalte | Go | v4.2 |
| **Site Server** (`site-server/`) | Minimaler HTTP/HTTPS-Webserver | Go (nur Stdlib) | v1.2 |

---

## Website

Single-File HTML5/CSS/JS – eine Datei, keine externen Abhängigkeiten, DSGVO-konform.

- Gruvbox Dark + Light Theme mit Toggle
- Terminal-Emulator-Design mit Hack-Font (Base64-eingebettet)
- Blog mit Multi-Kategorie-Support und Markdown XL Rendering
- Impressum und Datenschutzerklärung (DSGVO-konform)
- Online-Status mit farbiger Anzeige (grün/orange/rot)
- Alle Farben über Theme-System anpassbar (Dark + Light Mode)
- Wartungsmodus per config.json
- WCAG 2.1 AA Barrierefreiheit
- Alle Inhalte in JSON-Dateien externalisiert – ohne JSONs ist die Seite leer

### Schnellstart

1. Alle Dateien auf einen Webserver legen (oder den Site Server verwenden)
2. `config.json` mit eigenen Daten füllen
3. Fertig

---

## Site Manager

Go TUI-Anwendung im Midnight-Commander-Stil zur Verwaltung aller Website-Inhalte. Cross-Platform: Linux, macOS, Windows.

- Bearbeitung aller JSON-Dateien (Config, Blog, Impressum, Datenschutz)
- Theme-Editor mit 20-Farben-Palette und Hex-Eingabe (Dark + Light Mode, Gittermuster)
- SSH/SFTP-Profilverwaltung mit AES-256-GCM-verschlüsselten Passwörtern
- Status-Auswahl (online/abwesend/offline) mit farbiger Anzeige
- Ctrl+S Speichern-Shortcut in allen Formularen
- Meta-Description und Seitentitel werden direkt in der `index.html` aktualisiert (SEO)
- Copyright-Jahr und Datenschutz-Stand editierbar
- Wartungsmodus Ein/Aus-Schalter mit Vorschau

### Bauen und Starten

```bash
cd site-manager
make build                          # Aktuelle Plattform
./site-manager                      # Default: /var/www/html
./site-manager -root /pfad/zum/web  # Anderer Pfad
```

### Build-Targets

```bash
make build      # Aktuelle Plattform
make linux      # Linux amd64
make darwin     # macOS (ARM64 + Intel)
make windows    # Windows amd64 (.exe)
make all        # Alle Plattformen
make clean      # Aufräumen
```

### Bedienung

| Taste | Funktion |
|-------|----------|
| ↑↓ / Tab | Navigation |
| Ctrl+S | Speichern (überall) |
| Enter | Auswählen |
| Leertaste | Checkbox/Kategorie umschalten |
| Esc | Zurück |
| n / x | Neu / Löschen (in Listen) |
| q | Beenden |

Siehe [site-manager/README.md](site-manager/README.md) für Details.

---

## Site Server

Minimaler, abhängigkeitsfreier Webserver – ein einzelnes Go-Binary (nur Stdlib, ~5 MB). Ideal für minimale Server, Container und VMs.

- HTTP und HTTPS (optional, mit Let's Encrypt / Certbot)
- SPA-Support (Fallback auf `index.html`)
- Sicherheits-Header automatisch gesetzt
- Graceful TLS-Fallback: Platzhalter in der Config → Server startet im HTTP-Modus
- systemd-Integration mit interaktivem Installer (inkl. Dateibrowser für Zertifikate)
- Request-Logging via stdout / systemd-Journal

### Bauen und Installieren

```bash
cd site-server
make linux                # oder: make linux-arm64
sudo ./install.sh         # Interaktiver Installer
```

### Build-Targets

```bash
make build        # Aktuelle Plattform
make linux        # Linux amd64
make linux-arm64  # Linux arm64 (Raspberry Pi, ARM-Server)
make all          # Beide Architekturen
make clean        # Aufräumen
```

### config.json

```json
{
  "webroot": "/var/www/html",
  "port": 443,
  "cert_file": "/etc/letsencrypt/live/example.com/fullchain.pem",
  "key_file": "/etc/letsencrypt/live/example.com/privkey.pem"
}
```

Siehe [site-server/README.md](../site-server/README.md) für die vollständige TLS/Certbot-Anleitung.

---

## Dateistruktur

```
├── index.html              # Website (Single-File SPA)
├── config.json             # Persönliche Daten, Skills, Kontakt, Theme
├── blog.json               # Blog-Posts und Kategorien
├── impressum.json          # Impressum-Abschnitte
├── datenschutz.json        # Datenschutz-Abschnitte
├── CHANGELOG.md            # Versionshistorie (Website + Site Manager)
├── LICENSE                 # GPL-3.0
├── README.md               # Diese Datei
│
├── site-manager/           # Go TUI-Verwaltungstool
│   ├── main.go             # Hauptprogramm (~2100 Zeilen)
│   ├── go.mod              # Go-Modul + Dependencies
│   ├── Makefile            # Build-Targets (Linux, macOS, Windows)
│   ├── README.md
│   └── LICENSE             # GPL-3.0
│
└── site-server/            # Minimaler Webserver (separates Projekt)
    ├── main.go             # Server (~170 Zeilen, nur Go-Stdlib)
    ├── config.json         # Konfiguration (mit Platzhaltern)
    ├── install.sh          # Interaktiver Installer mit Dateibrowser
    ├── site-server.service.tpl  # systemd-Template
    ├── Makefile            # Build-Targets (Linux amd64/arm64)
    ├── CHANGELOG.md
    ├── README.md
    └── LICENSE             # GPL-3.0
```

## Konfiguration

### config.json

```json
{
  "name": "Max Mustermann",
  "seitentitel": "Meine Website",
  "beschreibung": "Seitenbeschreibung für SEO",
  "copyright_jahr": "2026",
  "jobtitle": "Entwickler",
  "branche": "IT",
  "bio": "Über mich...",
  "standort": "Berlin",
  "status": "online",
  "kontakt": {
    "email": "mail@example.com",
    "telefon": "+49123456789",
    "telefon_anzeige": "+49 123 456 789"
  },
  "skills": ["Go", "HTML", "CSS"],
  "impressum": {
    "name": "Max Mustermann",
    "strasse": "Musterstraße 1",
    "plz_ort": "12345 Berlin",
    "land": "Deutschland",
    "telefon": "+49 123 456 789",
    "email": "mail@example.com"
  },
  "terminal": {
    "user": "user",
    "host": "server"
  },
  "theme": {
    "bg": "#282828",
    "green": "#b8bb26",
    "text": "#ebdbb2",
    "grid_show": true,
    "grid_opacity": "0.25"
  }
}
```

### JSON-Platzhalter

In `impressum.json` und `datenschutz.json` können `{{impressum.*}}`-Platzhalter verwendet werden, die zur Laufzeit aus `config.json` aufgelöst werden:

- `{{impressum.name}}`, `{{impressum.strasse}}`, `{{impressum.plz_ort}}`
- `{{impressum.land}}`, `{{impressum.telefon}}`, `{{impressum.email}}`

## Releases

### Website (`index.html`)

| Version | Datum | Highlights |
|---------|-------|------------|
| v4.0 | 2. März 2026 | Theme-System (alle Farben anpassbar), Status-Farben (online/abwesend/offline), alle Fallback-Daten entfernt, Meta-Description + Copyright dynamisch |
| v3.0 | 1. März 2026 | Initiales Release – Terminal-Style SPA, Blog, Impressum, Datenschutz, Dark/Light Theme |

### Site Manager (`site-manager/`)

| Version | Datum | Highlights |
|---------|-------|------------|
| v4.2 | 2. März 2026 | Windows-Kompatibilität, `make windows`, `make all` |
| v4.1 | 2. März 2026 | Ctrl+S Speichern-Shortcut in allen Formularen |
| v4.0 | 2. März 2026 | Theme-Editor, SSH-Profile (AES-256-GCM), Status-Farben, Meta-Description, Copyright-Jahr, Datenschutz-Stand |
| v3.0 | 1. März 2026 | Initiales Release – TUI, Blog/Impressum/Datenschutz-Editor, SSH/SFTP, Wartungsmodus |

### Site Server (`site-server/`)

| Version | Datum | Highlights |
|---------|-------|------------|
| v1.2 | 3. März 2026 | Interaktiver TLS-Installer (Dateibrowser, Certbot-Erkennung), Graceful TLS-Fallback, Platzhalter in config.json |
| v1.0 | 2. März 2026 | Initiales Release – Minimaler HTTP/HTTPS-Server, SPA-Fallback, systemd-Integration, nur Go-Stdlib |

## Dependencies

### Site Manager

| Paket | Lizenz | Zweck |
|-------|--------|-------|
| github.com/rivo/tview | MIT | TUI-Framework |
| github.com/gdamore/tcell/v2 | Apache 2.0 | Terminal-Rendering |
| github.com/pkg/sftp | BSD-2 | SFTP-Dateitransfer |
| golang.org/x/crypto | BSD-3 | SSH-Verbindungen |

### Site Server

Keine externen Dependencies – ausschließlich Go-Stdlib.

Alle Dependencies sind GPL-3.0-kompatibel.

## Lizenz

GPL-3.0 – Siehe [LICENSE](LICENSE)
