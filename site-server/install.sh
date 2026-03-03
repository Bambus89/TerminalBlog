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
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    DIM='\033[2m'
    REVERSE='\033[7m'
    NC='\033[0m'
else
    GREEN='' YELLOW='' RED='' CYAN='' BOLD='' DIM='' REVERSE='' NC=''
fi

info()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; }
ask()   { echo -en "${BOLD}$1${NC}"; }

# ============================================================
# Dateibrowser – navigiert durch Verzeichnisse, wählt Dateien
# Parameter: $1 = Startverzeichnis, $2 = Dateityp-Filter (z.B. "*.pem")
# Gibt den gewählten Pfad auf stdout aus, leer bei Abbruch
# ============================================================
file_browser() {
    local current_dir="${1:-/}"
    local filter="${2:-*}"
    local selected=""

    while true; do
        # Verzeichnis auflösen
        current_dir="$(cd "$current_dir" 2>/dev/null && pwd)" || current_dir="/"

        echo ""
        echo -e "${BOLD}── Dateibrowser ──${NC}"
        echo -e "${DIM}Verzeichnis:${NC} ${CYAN}${current_dir}${NC}"
        echo -e "${DIM}Filter: ${filter} │ Eingabe: Nummer │ p = Pfad eingeben │ q = Abbrechen${NC}"
        echo ""

        # Einträge sammeln
        local entries=()
        local display=()
        local idx=0

        # Übergeordnetes Verzeichnis (immer als erstes)
        if [ "$current_dir" != "/" ]; then
            entries+=("..")
            display+=("${BOLD}📁 ../${NC} ${DIM}(übergeordnetes Verzeichnis)${NC}")
            idx=$((idx + 1))
        fi

        # Verzeichnisse auflisten
        while IFS= read -r -d '' dir; do
            local dirname
            dirname="$(basename "$dir")"
            entries+=("$dir")
            display+=("${BOLD}📁 ${dirname}/${NC}")
            idx=$((idx + 1))
        done < <(find "$current_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort -z)

        # Dateien auflisten (gefiltert)
        while IFS= read -r -d '' file; do
            local filename
            filename="$(basename "$file")"
            local filesize
            filesize="$(du -h "$file" 2>/dev/null | cut -f1)"
            entries+=("$file")
            display+=("📄 ${GREEN}${filename}${NC} ${DIM}(${filesize})${NC}")
            idx=$((idx + 1))
        done < <(find "$current_dir" -mindepth 1 -maxdepth 1 -type f -name "$filter" 2>/dev/null | sort -z)

        # Auch alle .pem/.crt/.key Dateien anzeigen wenn Filter *.pem ist
        if [ "$filter" = "*.pem" ]; then
            for ext in "*.crt" "*.key" "*.cer" "*.cert"; do
                while IFS= read -r -d '' file; do
                    # Duplikate vermeiden
                    local already=0
                    for e in "${entries[@]}"; do
                        if [ "$e" = "$file" ]; then already=1; break; fi
                    done
                    if [ "$already" -eq 0 ]; then
                        local filename
                        filename="$(basename "$file")"
                        local filesize
                        filesize="$(du -h "$file" 2>/dev/null | cut -f1)"
                        entries+=("$file")
                        display+=("📄 ${GREEN}${filename}${NC} ${DIM}(${filesize})${NC}")
                        idx=$((idx + 1))
                    fi
                done < <(find "$current_dir" -mindepth 1 -maxdepth 1 -type f -name "$ext" 2>/dev/null | sort -z)
            done
        fi

        # Anzeigen
        if [ ${#entries[@]} -eq 0 ]; then
            echo -e "  ${DIM}(keine Einträge)${NC}"
        else
            for i in "${!display[@]}"; do
                printf "  %s${BOLD}%2d${NC}%s  %b\n" "[" "$((i + 1))" "]" "${display[$i]}"
            done
        fi

        echo ""
        ask "Auswahl: "
        read -r choice

        # Abbruch
        if [[ "$choice" =~ ^[qQ]$ ]]; then
            echo ""
            return 1
        fi

        # Pfad manuell eingeben
        if [[ "$choice" =~ ^[pP]$ ]]; then
            ask "Pfad eingeben: "
            read -r manual_path
            if [ -n "$manual_path" ]; then
                if [ -d "$manual_path" ]; then
                    current_dir="$manual_path"
                    continue
                elif [ -f "$manual_path" ]; then
                    selected="$manual_path"
                    break
                else
                    warn "Pfad existiert nicht: $manual_path"
                    continue
                fi
            fi
            continue
        fi

        # Nummer prüfen
        if ! [[ "$choice" =~ ^[0-9]+$ ]]; then
            warn "Bitte eine Nummer, 'p' oder 'q' eingeben."
            continue
        fi

        local num=$((choice - 1))
        if [ "$num" -lt 0 ] || [ "$num" -ge "${#entries[@]}" ]; then
            warn "Ungültige Auswahl."
            continue
        fi

        local chosen="${entries[$num]}"

        # .. = übergeordnetes Verzeichnis
        if [ "$chosen" = ".." ]; then
            current_dir="$(dirname "$current_dir")"
            continue
        fi

        # Verzeichnis → hinein navigieren
        if [ -d "$chosen" ]; then
            current_dir="$chosen"
            continue
        fi

        # Datei → ausgewählt
        if [ -f "$chosen" ]; then
            selected="$chosen"
            break
        fi
    done

    echo -e "  ${GREEN}✓ Gewählt:${NC} $selected"
    echo "$selected"
}

# ============================================================
# TLS-Konfiguration abfragen
# ============================================================
configure_tls() {
    echo ""
    echo -e "${BOLD}── TLS-Konfiguration ──${NC}"
    echo ""
    echo "  [1]  Kein TLS (reiner HTTP-Server)"
    echo "  [2]  Let's Encrypt / Certbot (automatisch suchen)"
    echo "  [3]  Zertifikate manuell auswählen (Dateibrowser)"
    echo "  [4]  Pfade direkt eingeben"
    echo ""
    ask "Auswahl [1]: "
    read -r tls_choice
    tls_choice="${tls_choice:-1}"

    CERT_FILE=""
    KEY_FILE=""

    case "$tls_choice" in
        1)
            info "TLS deaktiviert – Server läuft im HTTP-Modus."
            ;;
        2)
            # Certbot-Zertifikate automatisch suchen
            echo ""
            local le_dir="/etc/letsencrypt/live"
            if [ ! -d "$le_dir" ]; then
                warn "Kein Let's-Encrypt-Verzeichnis gefunden ($le_dir)."
                warn "Bitte zuerst Certbot ausführen: sudo certbot certonly --standalone -d deine-domain.de"
                ask "Trotzdem Pfade manuell eingeben? [j/N]: "
                read -r manual_fallback
                if [[ "$manual_fallback" =~ ^[jJyY]$ ]]; then
                    configure_tls_manual
                fi
                return
            fi

            # Domains auflisten
            local domains=()
            while IFS= read -r -d '' domain_dir; do
                local domain
                domain="$(basename "$domain_dir")"
                # Nur Verzeichnisse mit fullchain.pem
                if [ -f "$domain_dir/fullchain.pem" ] && [ -f "$domain_dir/privkey.pem" ]; then
                    domains+=("$domain")
                fi
            done < <(find "$le_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort -z)

            if [ ${#domains[@]} -eq 0 ]; then
                warn "Keine Zertifikate in $le_dir gefunden."
                ask "Pfade manuell eingeben? [j/N]: "
                read -r manual_fallback
                if [[ "$manual_fallback" =~ ^[jJyY]$ ]]; then
                    configure_tls_manual
                fi
                return
            fi

            echo -e "${BOLD}Gefundene Let's-Encrypt-Zertifikate:${NC}"
            echo ""
            for i in "${!domains[@]}"; do
                local dom="${domains[$i]}"
                local cert_path="$le_dir/$dom/fullchain.pem"
                local expires=""
                if command -v openssl &>/dev/null; then
                    expires="$(openssl x509 -enddate -noout -in "$cert_path" 2>/dev/null | cut -d= -f2 || true)"
                fi
                if [ -n "$expires" ]; then
                    printf "  [%d]  %s ${DIM}(gültig bis: %s)${NC}\n" "$((i + 1))" "$dom" "$expires"
                else
                    printf "  [%d]  %s\n" "$((i + 1))" "$dom"
                fi
            done
            echo ""

            ask "Domain wählen [1]: "
            read -r dom_choice
            dom_choice="${dom_choice:-1}"

            local dom_idx=$((dom_choice - 1))
            if [ "$dom_idx" -ge 0 ] && [ "$dom_idx" -lt "${#domains[@]}" ]; then
                local chosen_domain="${domains[$dom_idx]}"
                CERT_FILE="$le_dir/$chosen_domain/fullchain.pem"
                KEY_FILE="$le_dir/$chosen_domain/privkey.pem"
                info "Zertifikat: $CERT_FILE"
                info "Schlüssel:  $KEY_FILE"
            else
                warn "Ungültige Auswahl – TLS deaktiviert."
            fi
            ;;
        3)
            # Dateibrowser für Zertifikat
            echo ""
            echo -e "${BOLD}Zertifikat auswählen (fullchain.pem):${NC}"
            local start_dir="/etc/letsencrypt/live"
            if [ ! -d "$start_dir" ]; then
                start_dir="/etc/ssl"
                if [ ! -d "$start_dir" ]; then
                    start_dir="/"
                fi
            fi

            local cert_result
            if cert_result="$(file_browser "$start_dir" "*.pem")"; then
                # Letzte Zeile ist der Pfad
                CERT_FILE="$(echo "$cert_result" | tail -1)"
            fi

            if [ -n "$CERT_FILE" ]; then
                echo ""
                echo -e "${BOLD}Schlüssel auswählen (privkey.pem):${NC}"
                # Im selben Verzeichnis starten wie das Zertifikat
                local key_start_dir
                key_start_dir="$(dirname "$CERT_FILE")"

                local key_result
                if key_result="$(file_browser "$key_start_dir" "*.pem")"; then
                    KEY_FILE="$(echo "$key_result" | tail -1)"
                fi
            fi

            if [ -z "$CERT_FILE" ] || [ -z "$KEY_FILE" ]; then
                warn "Auswahl abgebrochen – TLS deaktiviert."
                CERT_FILE=""
                KEY_FILE=""
            fi
            ;;
        4)
            configure_tls_manual
            ;;
        *)
            info "TLS deaktiviert."
            ;;
    esac
}

configure_tls_manual() {
    ask "Pfad zum Zertifikat (fullchain.pem): "
    read -r CERT_FILE
    if [ -n "$CERT_FILE" ]; then
        ask "Pfad zum Schlüssel (privkey.pem): "
        read -r KEY_FILE
        if [ -z "$KEY_FILE" ]; then
            error "Schlüssel ist Pflicht wenn ein Zertifikat angegeben wird."
            CERT_FILE=""
        fi
    fi
}

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

# TLS-Konfiguration (Menü mit Optionen)
CERT_FILE=""
KEY_FILE=""
configure_tls

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

# TLS-Zertifikate lesbar machen
if [ -n "$CERT_FILE" ]; then
    CERT_DIR="$(dirname "$CERT_FILE")"
    KEY_DIR="$(dirname "$KEY_FILE")"
    # ReadWritePaths um Zertifikatsverzeichnisse erweitern
    SERVICE_CONTENT="${SERVICE_CONTENT/ReadWritePaths={{WEBROOT}}/ReadWritePaths=$WEBROOT
ReadOnlyPaths=$CERT_DIR $KEY_DIR}"
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
