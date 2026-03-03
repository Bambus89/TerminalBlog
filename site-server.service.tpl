[Unit]
Description=Site Server – Minimaler Webserver
After=network.target
Wants=network-online.target

[Service]
Type=simple
ExecStart={{INSTALL_DIR}}/site-server -config {{INSTALL_DIR}}/config.json
WorkingDirectory={{INSTALL_DIR}}
Restart=on-failure
RestartSec=5

# Sicherheit
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths={{WEBROOT}}
PrivateTmp=true

# User (wird vom Installer gesetzt)
User={{USER}}
Group={{GROUP}}

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=site-server

[Install]
WantedBy=multi-user.target
