# Site Manager v4.2

TUI-Anwendung zur Verwaltung der Website-Inhalte (config.json, blog.json, impressum.json, datenschutz.json).
Unterstützt lokale Dateien und Remote-Server via SSH/SFTP. Cross-Platform: Linux, macOS, Windows.

## Features

- Konfiguration aller Website-Inhalte über ein Terminal-UI
- Theme-Editor mit 20-Farben-Palette und Hex-Eingabe (Dark + Light Mode, Gittermuster)
- SSH/SFTP-Profilverwaltung mit AES-256-GCM-verschlüsselten Passwörtern
- Blog-Editor mit Multi-Kategorie-Support und Farbauswahl
- Impressum- und Datenschutz-Editor mit Platzhalter-System
- Online-Status (online/abwesend/offline) mit farbiger Anzeige
- Meta-Description und Seitentitel werden direkt in der index.html aktualisiert
- Ctrl+S Speichern-Shortcut in allen Formularen
- Wartungsmodus mit Vorschau

## Bauen

```bash
make build      # Aktuelle Plattform
make linux      # Linux amd64
make darwin     # macOS (ARM64 + Intel)
make windows    # Windows amd64 (.exe)
make all        # Alle Plattformen
make clean      # Aufräumen
```

## Starten

```bash
./site-manager                        # Default: /var/www/html (Linux/macOS) bzw. .exe-Verzeichnis (Windows)
./site-manager -root /var/www/html    # Explizit
./site-manager -r /pfad/zum/web       # Kurzform
./site-manager /pfad/zum/web          # Als Argument
```

Der Document Root kann auch zur Laufzeit im Menü geändert werden.

## SSH/SFTP

Über "SSH/SFTP Verbindung" im Menü. Unterstützt:

- SSH-Key (mit optionaler Passphrase)
- Passwort (AES-256-GCM-verschlüsselt gespeichert)
- Profile speichern/laden/bearbeiten/löschen
- Master-Passwort für verschlüsselte Passwörter

## Bedienung

- **↑↓ / Tab / Shift+Tab** – Navigation
- **Ctrl+S** – Speichern (in allen Formularen)
- **Enter** – Auswählen / Button drücken
- **Leertaste** – Checkbox/Kategorie umschalten
- **Esc** – Zurück (überall)
- **n** – Neues Element (in Listen)
- **x** – Löschen (in Listen)
- **q** – Beenden (Hauptmenü)
- **Maus** – wird unterstützt

## Releases

| Version | Datum | Highlights |
|---------|-------|------------|
| v4.2 | 2. März 2026 | Windows-Kompatibilität, `make windows`, `make all` |
| v4.1 | 2. März 2026 | Ctrl+S Speichern-Shortcut in allen Formularen |
| v4.0 | 2. März 2026 | Theme-Editor, SSH-Profile, Status-Farben, Meta-Description, Copyright-Jahr, Datenschutz-Stand |
| v3.0 | 1. März 2026 | Initiales Release |

Siehe [CHANGELOG.md](../CHANGELOG.md) für Details.

## Lizenz

GPL-3.0 – siehe [LICENSE](LICENSE)
