#!/usr/bin/env bash
# ============================================================
# site-server Installer
# Distributionsunabhängig – benötigt: bash, systemd, sudo
# ============================================================
set -euo pipefail

VERSION="1.2.0"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY_NAME="site-server"

# Farben (falls Terminal unterstützt)
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    RED='\033[0;31m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    GREEN='' YELLOW='' RED='' BOLD='' NC=''
fi

info()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; }
ask()   { echo -en "${BOLD}$1${NC}"; }

# ============================================================
# Voraussetzungen prüfen
# ============================================================
if ! command -v systemctl &>/dev/null; then
    error "systemd wurde nicht gefunden. Dieser Installer benötigt systemd."
    exit 1
fi

# Binary finden
BINARY=""
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH_SUFFIX="linux-amd64" ;;
    aarch64) ARCH_SUFFIX="linux-arm64" ;;
    *)       ARCH_SUFFIX="" ;;
esac

# Zuerst architekturspezifisches Binary suchen, dann generisches
if [ -n "$ARCH_SUFFIX" ] && [ -f "$SCRIPT_DIR/${BINARY_NAME}-${ARCH_SUFFIX}" ]; then
    BINARY="$SCRIPT_DIR/${BINARY_NAME}-${ARCH_SUFFIX}"
elif [ -f "$SCRIPT_DIR/$BINARY_NAME" ]; then
    BINARY="$SCRIPT_DIR/$BINARY_NAME"
else
    error "Kein site-server Binary gefunden in: $SCRIPT_DIR"
    echo "  Bitte zuerst kompilieren: make build (oder make linux)"
    exit 1
fi

# ============================================================
# Konfiguration abfragen
# ============================================================
echo ""
echo -e "${BOLD}═══ site-server v${VERSION} – Installer ═══${NC}"
echo ""

# Installationsverzeichnis
ask "Installationsverzeichnis [/opt/site-server]: "
read -r INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-/opt/site-server}"
INSTALL_DIR="${INSTALL_DIR%/}"  # Trailing Slash entfernen

# WebRoot
ask "WebRoot (Verzeichnis mit index.html) [/var/www/html]: "
read -r WEBROOT
WEBROOT="${WEBROOT:-/var/www/html}"
WEBROOT="${WEBROOT%/}"

# Port
ask "Port [8080]: "
read -r PORT
PORT="${PORT:-8080}"

# Nur Zahlen prüfen
if ! [[ "$PORT" =~ ^[0-9]+$ ]]; then
    error "Port muss eine Zahl sein."
    exit 1
fi

# TLS-Zertifikate (optional)
ask "TLS-Zertifikat (leer = kein TLS): "
read -r CERT_FILE

KEY_FILE=""
if [ -n "$CERT_FILE" ]; then
    ask "TLS-Schlüssel: "
    read -r KEY_FILE
    if [ -z "$KEY_FILE" ]; then
        error "TLS-Schlüssel ist Pflicht wenn ein Zertifikat angegeben wird."
        exit 1
    fi
fi

# Benutzer
CURRENT_USER="$(whoami)"
ask "Service-Benutzer [$CURRENT_USER]: "
read -r SVC_USER
SVC_USER="${SVC_USER:-$CURRENT_USER}"

# Gruppe
DEFAULT_GROUP="$(id -gn "$SVC_USER" 2>/dev/null || echo "$SVC_USER")"
ask "Service-Gruppe [$DEFAULT_GROUP]: "
read -r SVC_GROUP
SVC_GROUP="${SVC_GROUP:-$DEFAULT_GROUP}"

# Service-Name
ask "Service-Name [site-server]: "
read -r SVC_NAME
SVC_NAME="${SVC_NAME:-site-server}"

# ============================================================
# Zusammenfassung
# ============================================================
echo ""
echo -e "${BOLD}── Zusammenfassung ──${NC}"
echo "  Binary:       $BINARY"
echo "  Installation: $INSTALL_DIR"
echo "  WebRoot:      $WEBROOT"
echo "  Port:         $PORT"
if [ -n "$CERT_FILE" ]; then
    echo "  TLS Cert:     $CERT_FILE"
    echo "  TLS Key:      $KEY_FILE"
else
    echo "  TLS:          deaktiviert"
fi
echo "  Benutzer:     $SVC_USER:$SVC_GROUP"
echo "  Service:      $SVC_NAME"
echo ""

ask "Installation starten? [j/N]: "
read -r CONFIRM
if [[ ! "$CONFIRM" =~ ^[jJyY]$ ]]; then
    warn "Abgebrochen."
    exit 0
fi

# ============================================================
# Installation
# ============================================================
echo ""

# Verzeichnis erstellen
sudo mkdir -p "$INSTALL_DIR"
info "Verzeichnis erstellt: $INSTALL_DIR"

# Binary kopieren
sudo cp "$BINARY" "$INSTALL_DIR/$BINARY_NAME"
sudo chmod 755 "$INSTALL_DIR/$BINARY_NAME"
info "Binary installiert: $INSTALL_DIR/$BINARY_NAME"

# Config erstellen (nur wenn noch keine existiert)
CONFIG_PATH="$INSTALL_DIR/config.json"
if [ -f "$CONFIG_PATH" ]; then
    warn "config.json existiert bereits – wird nicht überschrieben."
    warn "Aktuelle Config wird beibehalten. Bitte manuell anpassen falls nötig."
else
    # JSON erstellen
    CONFIG_JSON=$(cat <<JSONEOF
{
  "webroot": "$WEBROOT",
  "port": $PORT,
  "cert_file": "$CERT_FILE",
  "key_file": "$KEY_FILE"
}
JSONEOF
)
    echo "$CONFIG_JSON" | sudo tee "$CONFIG_PATH" > /dev/null
    sudo chmod 644 "$CONFIG_PATH"
    info "Config erstellt: $CONFIG_PATH"
fi

# WebRoot erstellen (falls nicht vorhanden)
if [ ! -d "$WEBROOT" ]; then
    sudo mkdir -p "$WEBROOT"
    sudo chown "$SVC_USER:$SVC_GROUP" "$WEBROOT"
    info "WebRoot erstellt: $WEBROOT"
else
    info "WebRoot existiert bereits: $WEBROOT"
fi

# Ownership setzen
sudo chown -R "$SVC_USER:$SVC_GROUP" "$INSTALL_DIR"
info "Besitzer gesetzt: $SVC_USER:$SVC_GROUP"

# ============================================================
# systemd Service installieren
# ============================================================
SERVICE_FILE="/etc/systemd/system/${SVC_NAME}.service"

# Template verarbeiten
if [ -f "$SCRIPT_DIR/site-server.service.tpl" ]; then
    TEMPLATE="$(cat "$SCRIPT_DIR/site-server.service.tpl")"
else
    # Fallback: Template inline
    TEMPLATE='[Unit]
Description=Site Server – Minimaler Webserver
After=network.target
Wants=network-online.target

[Service]
Type=simple
ExecStart={{INSTALL_DIR}}/site-server -config {{INSTALL_DIR}}/config.json
WorkingDirectory={{INSTALL_DIR}}
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths={{WEBROOT}}
PrivateTmp=true
User={{USER}}
Group={{GROUP}}
StandardOutput=journal
StandardError=journal
SyslogIdentifier=site-server

[Install]
WantedBy=multi-user.target'
fi

# Platzhalter ersetzen
SERVICE_CONTENT="${TEMPLATE//\{\{INSTALL_DIR\}\}/$INSTALL_DIR}"
SERVICE_CONTENT="${SERVICE_CONTENT//\{\{WEBROOT\}\}/$WEBROOT}"
SERVICE_CONTENT="${SERVICE_CONTENT//\{\{USER\}\}/$SVC_USER}"
SERVICE_CONTENT="${SERVICE_CONTENT//\{\{GROUP\}\}/$SVC_GROUP}"

# Port < 1024 → braucht Capability
if [ "$PORT" -lt 1024 ]; then
    SERVICE_CONTENT="${SERVICE_CONTENT/\[Service\]/[Service]
AmbientCapabilities=CAP_NET_BIND_SERVICE}"
    info "Port $PORT < 1024 → CAP_NET_BIND_SERVICE wird gesetzt"
fi

echo "$SERVICE_CONTENT" | sudo tee "$SERVICE_FILE" > /dev/null
sudo chmod 644 "$SERVICE_FILE"
info "Service installiert: $SERVICE_FILE"

# systemd neu laden und Service aktivieren
sudo systemctl daemon-reload
sudo systemctl enable "$SVC_NAME"
info "Service aktiviert: $SVC_NAME"

# ============================================================
# Service starten
# ============================================================
ask "Service jetzt starten? [j/N]: "
read -r START_NOW
if [[ "$START_NOW" =~ ^[jJyY]$ ]]; then
    sudo systemctl start "$SVC_NAME"
    sleep 1
    if systemctl is-active --quiet "$SVC_NAME"; then
        info "Service läuft!"
    else
        error "Service konnte nicht gestartet werden."
        echo "  Logs prüfen: sudo journalctl -u $SVC_NAME -f"
    fi
else
    info "Service nicht gestartet. Manuell starten mit:"
    echo "  sudo systemctl start $SVC_NAME"
fi

# ============================================================
# Fertig
# ============================================================
echo ""
echo -e "${BOLD}── Installation abgeschlossen ──${NC}"
echo ""
echo "  Nützliche Befehle:"
echo "    sudo systemctl start $SVC_NAME      # Starten"
echo "    sudo systemctl stop $SVC_NAME       # Stoppen"
echo "    sudo systemctl restart $SVC_NAME    # Neustarten"
echo "    sudo systemctl status $SVC_NAME     # Status"
echo "    sudo journalctl -u $SVC_NAME -f     # Logs"
echo ""
echo "  Config: $INSTALL_DIR/config.json"
echo "  Binary: $INSTALL_DIR/$BINARY_NAME"
echo ""

PROTOCOL="http"
if [ -n "$CERT_FILE" ]; then
    PROTOCOL="https"
fi
info "Erreichbar unter: ${PROTOCOL}://localhost:${PORT}"
echo ""
