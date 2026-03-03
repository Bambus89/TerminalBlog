# Personal Terminal Website + Site Manager

Eine moderne, responsive Einzelseiten-Website im Terminal-Stil mit Gruvbox-Theme, Blog-System und einer TUI-Verwaltungsoberfläche.

## Features

### Website (`index.html`)
- Single-File HTML5/CSS/JS – keine externen Abhängigkeiten
- Gruvbox Dark + Light Theme mit Toggle
- Terminal-Emulator-Design mit Hack-Font (Base64-eingebettet, DSGVO-konform)
- Blog mit Multi-Kategorie-Support und Markdown XL Rendering
- Impressum und Datenschutzerklärung (DSGVO-konform)
- Online-Status mit farbiger Anzeige (grün/orange/rot)
- Wartungsmodus per config.json
- Alle Farben über Theme-System anpassbar (Dark + Light Mode)
- WCAG 2.1 AA Barrierefreiheit
- Alle Inhalte in JSON-Dateien externalisiert – ohne JSONs ist die Seite leer

### Site Manager (`site-manager/`)
- Go TUI-Anwendung im Midnight-Commander-Stil
- Bearbeitung aller JSON-Dateien (Config, Blog, Impressum, Datenschutz)
- **Theme-Editor** mit Farbpalette und Hex-Eingabe (20+ Farben, Dark + Light Mode, Gittermuster)
- **SSH/SFTP-Profilverwaltung** mit AES-256-GCM-verschlüsselten Passwörtern und Master-Passwort
- **Status-Auswahl** (online/abwesend/offline) mit farbiger Anzeige auf der Website
- **Ctrl+S Speichern-Shortcut** in allen Formularen (angezeigt in der Statusbar)
- Meta-Description und Seitentitel werden direkt in der `index.html` aktualisiert (SEO)
- Copyright-Jahr und Datenschutz-Stand editierbar
- Farbpalette für Blog-Kategorie-Farben
- Wartungsmodus Ein/Aus-Schalter mit Vorschau
- Document Root zur Laufzeit änderbar
- Cross-Platform: Linux, macOS, Windows

## Schnellstart

### Website

1. Alle Dateien auf einen Webserver legen (oder den [site-server](../site-server/) verwenden)
2. `config.json` mit eigenen Daten füllen
3. Fertig – die Website läuft

### Site Manager bauen und starten

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

## Dateistruktur

```
├── index.html          # Website (Single-File SPA)
├── config.json         # Persönliche Daten, Skills, Kontakt, Theme
├── blog.json           # Blog-Posts und Kategorien
├── impressum.json      # Impressum-Abschnitte
├── datenschutz.json    # Datenschutz-Abschnitte (§1-§13)
├── CHANGELOG.md        # Versionshistorie
├── LICENSE             # GPL-3.0
├── README.md           # Diese Datei
└── site-manager/       # Go TUI-Verwaltungstool
    ├── main.go         # Hauptprogramm (~2100 Zeilen)
    ├── go.mod          # Go-Modul + Dependencies
    ├── Makefile         # Build-Targets (Linux, macOS, Windows)
    └── ssh.json        # SSH-Profile (wird automatisch erstellt, 0600)
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

## Dependencies

| Paket | Lizenz | Zweck |
|-------|--------|-------|
| github.com/rivo/tview | MIT | TUI-Framework |
| github.com/gdamore/tcell/v2 | Apache 2.0 | Terminal-Rendering |
| github.com/pkg/sftp | BSD-2 | SFTP-Dateitransfer |
| golang.org/x/crypto | BSD-3 | SSH-Verbindungen |

Alle Dependencies sind GPL-3.0-kompatibel.

## Lizenz

GPL-3.0 – Siehe [LICENSE](LICENSE)
