package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const version = "1.2.0"

type Config struct {
	WebRoot  string `json:"webroot"`
	Port     int    `json:"port"`
	CertFile string `json:"cert_file,omitempty"`
	KeyFile  string `json:"key_file,omitempty"`
}

func defaultConfig() Config {
	return Config{
		WebRoot: "/var/www/html",
		Port:    8080,
	}
}

func loadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config ungültig: %w", err)
	}
	return cfg, nil
}

// securityHeaders setzt grundlegende Sicherheits-Header
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// loggingHandler loggt jede Anfrage
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s %s",
			r.Method, r.URL.Path, rec.statusCode,
			time.Since(start).Round(time.Microsecond), r.RemoteAddr)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// spaHandler serviert statische Dateien und fällt auf index.html zurück (SPA)
type spaHandler struct {
	root http.Dir
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pfad bereinigen
	path := filepath.Clean(r.URL.Path)
	if path == "." {
		path = "/"
	}

	// Verzeichnisauflistung verhindern
	if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "/" {
		// Prüfen ob index.html im Verzeichnis existiert
		indexPath := filepath.Join(string(h.root), path, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	// Datei suchen
	fullPath := filepath.Join(string(h.root), path)
	info, err := os.Stat(fullPath)

	if err != nil || info.IsDir() {
		// Datei nicht gefunden oder ist Verzeichnis → SPA Fallback
		if _, err := os.Stat(filepath.Join(string(h.root), "index.html")); err == nil {
			http.ServeFile(w, r, filepath.Join(string(h.root), "index.html"))
			return
		}
		http.NotFound(w, r)
		return
	}

	// Korrekte MIME-Types für Web-Assets
	http.ServeFile(w, r, fullPath)
}

func main() {
	configPath := flag.String("config", "config.json", "Pfad zur config.json")
	showVersion := flag.Bool("version", false, "Version anzeigen")
	flag.BoolVar(showVersion, "v", false, "Version anzeigen (Kurzform)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("site-server v%s\n", version)
		os.Exit(0)
	}

	// Config laden
	cfg, err := loadConfig(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Keine config.json gefunden, verwende Defaults (Port %d, WebRoot %s)",
				cfg.Port, cfg.WebRoot)
		} else {
			log.Fatalf("Fehler beim Laden der Config: %v", err)
		}
	}

	// WebRoot prüfen
	absRoot, err := filepath.Abs(cfg.WebRoot)
	if err != nil {
		log.Fatalf("Ungültiger WebRoot-Pfad: %v", err)
	}
	cfg.WebRoot = absRoot

	if info, err := os.Stat(cfg.WebRoot); err != nil || !info.IsDir() {
		log.Fatalf("WebRoot existiert nicht oder ist kein Verzeichnis: %s", cfg.WebRoot)
	}

	// Handler aufbauen
	handler := &spaHandler{root: http.Dir(cfg.WebRoot)}
	var h http.Handler = handler
	h = securityHeaders(h)
	h = loggingHandler(h)

	addr := fmt.Sprintf(":%d", cfg.Port)
	tls := cfg.CertFile != "" && cfg.KeyFile != ""

	// TLS-Zertifikate prüfen – bei Platzhaltern oder fehlenden Dateien auf HTTP zurückfallen
	if tls {
		certOK := true
		if _, err := os.Stat(cfg.CertFile); err != nil {
			log.Printf("WARNUNG: Zertifikat nicht gefunden: %s", cfg.CertFile)
			certOK = false
		}
		if _, err := os.Stat(cfg.KeyFile); err != nil {
			log.Printf("WARNUNG: Schlüssel nicht gefunden: %s", cfg.KeyFile)
			certOK = false
		}
		if !certOK {
			log.Printf("TLS deaktiviert – Zertifikate nicht vorhanden. Starte im HTTP-Modus.")
			log.Printf("Bitte cert_file und key_file in der config.json anpassen.")
			tls = false
		}
	}

	// Startup-Info
	protocol := "http"
	if tls {
		protocol = "https"
	}
	log.Printf("site-server v%s", version)
	log.Printf("WebRoot:  %s", cfg.WebRoot)
	log.Printf("Adresse:  %s://0.0.0.0:%d", protocol, cfg.Port)
	if tls {
		log.Printf("TLS:      %s / %s", cfg.CertFile, cfg.KeyFile)
	}

	// Server starten
	server := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if tls {
		log.Fatal(server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile))
	} else {
		log.Fatal(server.ListenAndServe())
	}
}
