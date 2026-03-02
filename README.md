# Personal Terminal Website + Site Manager

Eine moderne, responsive Einzelseiten-Website im Terminal-Stil mit Gruvbox-Theme, Blog-System und einer TUI-Verwaltungsoberfläche.

## Features

### Website (`index.html`)
- Single-File HTML5/CSS/JS – keine externen Abhängigkeiten
- Gruvbox Dark + Light Theme mit Toggle
- Terminal-Emulator-Design mit Hack-Font (Base64-eingebettet, DSGVO-konform)
- Blog mit Multi-Kategorie-Support und Markdown XL Rendering
- Impressum und Datenschutzerklärung (DSGVO-konform)
- Wartungsmodus per config.json
- WCAG 2.1 AA Barrierefreiheit
- Alle Inhalte in JSON-Dateien externalisiert

### Site Manager (`site-manager/`)
- Go TUI-Anwendung im Midnight-Commander-Stil
- Bearbeitung aller JSON-Dateien (Config, Blog, Impressum, Datenschutz)
- Theme-Editor mit Farbpalette und Hex-Eingabe (Dark + Light Mode)
- SSH/SFTP-Profilverwaltung mit AES-256-verschlüsselten Passwörtern
- Status-Auswahl (online/abwesend/offline) mit farbiger Anzeige
- Meta-Description und Seitentitel werden direkt in der HTML aktualisiert
- Datenschutz-Stand editierbar
- Farbpalette für Kategorie-Farben
- Wartungsmodus Ein/Aus-Schalter
- Document Root zur Laufzeit änderbar

## Schnellstart

1. Alle Dateien auf einen Webserver legen
2. `config.json` mit eigenen Daten füllen
3. Fertig – die Website läuft

### Site Manager bauen

```bash
cd site-manager
make build        # Aktuelle Plattform
make linux        # Linux amd64
make darwin       # macOS (ARM64 + AMD64)
```

### Site Manager starten

```bash
./site-manager                      # Default: /var/www/html
./site-manager -root /pfad/zum/web  # Anderer Pfad
```

## Dateistruktur

```
├── index.html          # Website (Single-File SPA)
├── config.json         # Persönliche Daten, Skills, Kontakt
├── blog.json           # Blog-Posts und Kategorien
├── impressum.json      # Impressum-Abschnitte
├── datenschutz.json    # Datenschutz-Abschnitte (§1-§13)
├── LICENSE             # GPL-3.0
└── site-manager/       # Go TUI-Verwaltungstool
    ├── main.go
    ├── go.mod
    ├── Makefile
    └── LICENSE         # GPL-3.0
```

## JSON-Platzhalter

In `impressum.json` und `datenschutz.json` können `{{impressum.*}}`-Platzhalter verwendet werden, die zur Laufzeit aus `config.json` aufgelöst werden:

- `{{impressum.name}}`, `{{impressum.strasse}}`, `{{impressum.plz_ort}}`
- `{{impressum.land}}`, `{{impressum.telefon}}`, `{{impressum.email}}`

## Lizenz

GPL-3.0 – siehe [LICENSE](LICENSE)
