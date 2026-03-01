# Site Manager v3.0

TUI-Anwendung zur Verwaltung der Website-Inhalte (config.json, blog.json, impressum.json, datenschutz.json).
Unterstützt lokale Dateien und Remote-Server via SSH/SFTP.

## Bauen

```bash
make build    # Aktuelle Plattform
make linux    # Linux amd64
make darwin   # macOS (ARM64 + AMD64)
```

## Starten

```bash
./site-manager                        # Default: /var/www/html
./site-manager -root /var/www/html    # Explizit
./site-manager -r /pfad/zum/web       # Kurzform
./site-manager /pfad/zum/web          # Als Argument
```

Der Document Root kann auch zur Laufzeit im Menü geändert werden.

## SSH/SFTP

Über "SSH/SFTP Verbindung" im Menü. Unterstützt:

- SSH-Key (mit optionaler Passphrase)
- Passwort

## Bedienung

- **↑↓ / Tab / Shift+Tab** – Navigation
- **Enter** – Auswählen / Button drücken
- **Leertaste** – Checkbox/Kategorie umschalten
- **Esc** – Zurück (überall)
- **n** – Neues Element (in Listen)
- **x** – Löschen (in Listen)
- **q** – Beenden (Hauptmenü)
- **Maus** – wird unterstützt

## Lizenz

GPL-3.0 – siehe [LICENSE](LICENSE)
