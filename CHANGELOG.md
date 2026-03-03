# Changelog

## v4.2 – 2. März 2026

### Neue Features

- **Windows-Kompatibilität** – Der Site-Manager kann jetzt auch für Windows kompiliert werden. Der Standard-Document-Root wird OS-abhängig gesetzt: unter Linux/macOS `/var/www/html`, unter Windows das Verzeichnis der `.exe`. Alle Pfadoperationen nutzen `filepath` für plattformübergreifende Kompatibilität.
- **Makefile: `make windows` und `make all`** – Neues Build-Target `windows` für Cross-Compile nach Windows (amd64). `make all` baut alle Plattformen (Linux, macOS, Windows) auf einmal.

### Build-Anleitung

```bash
make build              # Aktuelle Plattform
make linux              # Linux amd64
make darwin             # macOS (ARM64 + Intel)
make windows            # Windows amd64 (.exe)
make all                # Alle Plattformen
make clean              # Build-Artefakte aufräumen
```

## v4.1 – 2. März 2026

### Neue Features

- **Ctrl+S Speichern-Shortcut** – In allen Formularen und Editoren kann jetzt mit `Ctrl+S` gespeichert werden, ohne zum Speichern-Button tabben zu müssen. Der Shortcut wird in der Statusbar angezeigt. Funktioniert auch in Dialogen mit Tab-Cycling (Blog-Editor, Kategorie-Editor, Theme-Farbwähler).

## v4.0 – 2. März 2026

### Neue Features

- **Theme-Editor** – Neuer Menüpunkt "🎨 Theme" im Site-Manager. Alle Farben der Website (Dark + Light Mode) können über eine Farbpalette oder Hex-Eingabe angepasst werden. Das Hintergrund-Gittermuster lässt sich ein-/ausschalten und in der Opacity regulieren. Zurücksetzen auf Gruvbox-Standard mit einem Klick.
- **SSH-Profilverwaltung** – SSH/SFTP-Verbindungen können als Profile gespeichert und geladen werden. Passwörter werden mit AES-256-GCM verschlüsselt in einer `ssh.json` abgelegt (Dateiberechtigungen 0600). Beim Verbinden wird nur das Master-Passwort abgefragt. Profile können bearbeitet und gelöscht werden.
- **Status mit Farben** – Der Online-Status ist jetzt ein Auswahlmenü (online/abwesend/offline) statt Freitext. Auf der Website blinkt der Punkt grün (online), orange (abwesend) oder rot (offline).
- **Meta-Description dynamisch** – Das `<meta name="description">`-Tag wird aus `config.json` gelesen und vom Site-Manager direkt in der `index.html` aktualisiert (SEO-relevant).
- **Copyright-Jahr konfigurierbar** – Das Jahr im Copyright-Footer ist über `config.json` frei wählbar.
- **Datenschutz-Stand editierbar** – Der Stand der Datenschutzerklärung kann im Site-Manager geändert werden (Menüpunkt "📅 Stand" in der Datenschutz-Ansicht).

### Verbesserungen

- Site-Manager schreibt beim Speichern der Konfiguration auch `<meta description>` und `<title>` direkt in die `index.html` (nicht mehr nur zur JS-Laufzeit).
- Alle Fallback-Daten in der `index.html` entfernt – die Seite ist ohne JSON-Dateien leer, alle Inhalte kommen ausschließlich aus den JSONs.
- Keine personenbezogenen Daten mehr in der HTML-Datei.

### Konfiguration

Neue Felder in `config.json`:

```json
{
  "beschreibung": "Seitenbeschreibung für SEO",
  "copyright_jahr": "2026",
  "theme": {
    "bg": "#282828",
    "green": "#b8bb26",
    "text": "#ebdbb2",
    "grid_show": true,
    "grid_opacity": "0.25",
    "light_bg": "#fbf1c7",
    "light_green": "#79740e",
    "light_text": "#3c3836"
  }
}
```

## v3.0 – 1. März 2026

### Initiales Release

- Terminal-Style Website mit Gruvbox Dark/Light Theme
- Blog-System mit Multi-Kategorie-Support und Markdown XL
- Impressum und Datenschutzerklärung (DSGVO-konform)
- Go TUI Site-Manager mit SSH/SFTP-Unterstützung
- Alle Inhalte in JSON-Dateien externalisiert
- Wartungsmodus per Konfiguration
- GPL-3.0 Lizenz
