package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/sftp"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
)

// ===== Datenstrukturen =====

type Config struct {
	Name         string          `json:"name"`
	Initials     string          `json:"initials"`
	Seitentitel  string          `json:"seitentitel,omitempty"`
	JobTitle     string          `json:"jobtitle"`
	Branche      string          `json:"branche"`
	Bio          string          `json:"bio"`
	Standort     string          `json:"standort"`
	Status       string          `json:"status"`
	Wartung      bool            `json:"wartung"`
	WartungText  string          `json:"wartung_text,omitempty"`
	Kontakt      ConfigKontakt   `json:"kontakt"`
	Skills       []string        `json:"skills"`
	Impressum    ConfigImpressum `json:"impressum"`
	Terminal     ConfigTerminal  `json:"terminal"`
}

type ConfigKontakt struct {
	Email          string `json:"email"`
	Telefon        string `json:"telefon"`
	TelefonAnzeige string `json:"telefon_anzeige"`
}

type ConfigImpressum struct {
	Name    string `json:"name"`
	Strasse string `json:"strasse"`
	PlzOrt  string `json:"plz_ort"`
	Land    string `json:"land"`
	Telefon string `json:"telefon"`
	Email   string `json:"email"`
}

type ConfigTerminal struct {
	User string `json:"user"`
	Host string `json:"host"`
}

type BlogData struct {
	Categories []BlogCategory `json:"categories"`
	Posts      []BlogPost     `json:"posts"`
}

type BlogCategory struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type BlogPost struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Date     string      `json:"date"`
	Category interface{} `json:"category"` // string oder []string
	Tags     []string    `json:"tags"`
	Excerpt  string      `json:"excerpt"`
	Content  string      `json:"content"`
}

func (p BlogPost) GetCategories() []string {
	switch v := p.Category.(type) {
	case string:
		return []string{v}
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{}
}

func (p *BlogPost) SetCategories(cats []string) {
	if len(cats) == 1 {
		p.Category = cats[0]
	} else {
		p.Category = cats
	}
}

type LegalData struct {
	Titel      string         `json:"titel"`
	Stand      string         `json:"stand,omitempty"`
	Abschnitte []LegalSection `json:"abschnitte"`
}

type LegalSection struct {
	Nr           string `json:"nr,omitempty"`
	Ueberschrift string `json:"ueberschrift"`
	Inhalt       string `json:"inhalt"`
}

// ===== Connection Mode =====

type ConnMode int

const (
	ConnLocal ConnMode = iota
	ConnSSH
)

type SSHConnection struct {
	Host       string
	Port       string
	User       string
	AuthMethod string // "password" oder "key"
	Password   string
	KeyPath    string
	RemotePath string

	client     *ssh.Client
	sftpClient *sftp.Client
}

func (s *SSHConnection) Connect() error {
	var authMethods []ssh.AuthMethod

	if s.AuthMethod == "key" {
		keyData, err := os.ReadFile(s.KeyPath)
		if err != nil {
			return fmt.Errorf("SSH-Key lesen: %w", err)
		}
		var signer ssh.Signer
		if s.Password != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(s.Password))
		} else {
			signer, err = ssh.ParsePrivateKey(keyData)
		}
		if err != nil {
			return fmt.Errorf("SSH-Key parsen: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		authMethods = append(authMethods, ssh.Password(s.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(s.Host, s.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH-Verbindung: %w", err)
	}
	s.client = client

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return fmt.Errorf("SFTP-Session: %w", err)
	}
	s.sftpClient = sftpClient
	return nil
}

func (s *SSHConnection) Close() {
	if s.sftpClient != nil {
		s.sftpClient.Close()
	}
	if s.client != nil {
		s.client.Close()
	}
}

func (s *SSHConnection) ReadFile(path string) ([]byte, error) {
	f, err := s.sftpClient.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (s *SSHConnection) WriteFile(path string, data []byte) error {
	f, err := s.sftpClient.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func (s *SSHConnection) StatDir(path string) (bool, error) {
	info, err := s.sftpClient.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ===== Globale Variablen =====

var (
	docRoot  string
	connMode ConnMode
	sshConn  *SSHConnection

	app   *tview.Application
	pages *tview.Pages

	config      Config
	blog        BlogData
	impressum   LegalData
	datenschutz LegalData
)

// ===== JSON I/O (lokal + SSH) =====

func jsonPath(filename string) string {
	if connMode == ConnSSH && sshConn != nil {
		return sshConn.RemotePath + "/" + filename
	}
	return filepath.Join(docRoot, filename)
}

func readFileBytes(path string) ([]byte, error) {
	if connMode == ConnSSH && sshConn != nil {
		return sshConn.ReadFile(path)
	}
	return os.ReadFile(path)
}

func writeFileBytes(path string, data []byte) error {
	if connMode == ConnSSH && sshConn != nil {
		return sshConn.WriteFile(path, data)
	}
	return os.WriteFile(path, data, 0644)
}

func loadJSON(filename string, target interface{}) error {
	data, err := readFileBytes(jsonPath(filename))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func saveJSON(filename string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return writeFileBytes(jsonPath(filename), bytes)
}

func loadAll() {
	if err := loadJSON("config.json", &config); err != nil {
		config = Config{Name: "Neue Website", Status: "online"}
	}
	if err := loadJSON("blog.json", &blog); err != nil {
		blog = BlogData{}
	}
	if err := loadJSON("impressum.json", &impressum); err != nil {
		impressum = LegalData{Titel: "Impressum"}
	}
	if err := loadJSON("datenschutz.json", &datenschutz); err != nil {
		datenschutz = LegalData{Titel: "Datenschutzerklärung"}
	}
}

// ===== Styling =====

var (
	colorBg      = tcell.NewRGBColor(40, 40, 40)
	colorBgPanel = tcell.NewRGBColor(29, 32, 33)
	colorBgCard  = tcell.NewRGBColor(60, 56, 54)
	colorGreen   = tcell.NewRGBColor(184, 187, 38)
	colorText    = tcell.NewRGBColor(235, 219, 178)
	colorTextDim = tcell.NewRGBColor(168, 153, 132)
	colorYellow  = tcell.NewRGBColor(250, 189, 47)
	colorCyan    = tcell.NewRGBColor(131, 165, 152)
	colorOrange  = tcell.NewRGBColor(254, 128, 25)
	colorRed     = tcell.NewRGBColor(251, 73, 52)
)

func styledList() *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(colorBg)
	list.SetMainTextColor(colorText)
	list.SetSecondaryTextColor(colorTextDim)
	list.SetSelectedTextColor(colorBg)
	list.SetSelectedBackgroundColor(colorGreen)
	list.SetBorder(true)
	list.SetBorderColor(colorBgCard)
	list.SetTitleColor(colorGreen)
	return list
}

func styledForm() *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colorBg)
	form.SetFieldBackgroundColor(colorBgCard)
	form.SetFieldTextColor(colorText)
	form.SetLabelColor(colorCyan)
	// Inaktive Buttons: dezenter Hintergrund, grüner Text
	form.SetButtonBackgroundColor(colorBgCard)
	form.SetButtonTextColor(colorGreen)
	// Aktiver/fokussierter Button: grüner Hintergrund, dunkler Text
	form.SetButtonActivatedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.NewRGBColor(184, 187, 38)))
	form.SetBorder(true)
	form.SetBorderColor(colorBgCard)
	form.SetTitleColor(colorGreen)
	return form
}

func styledTextView() *tview.TextView {
	tv := tview.NewTextView()
	tv.SetBackgroundColor(colorBg)
	tv.SetTextColor(colorText)
	tv.SetBorder(true)
	tv.SetBorderColor(colorBgCard)
	tv.SetTitleColor(colorGreen)
	tv.SetDynamicColors(true)
	return tv
}

func statusBar(text string) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetBackgroundColor(colorBgPanel)
	tv.SetTextColor(colorTextDim)
	tv.SetTextAlign(tview.AlignCenter)
	tv.SetText(text)
	tv.SetDynamicColors(true)
	return tv
}

// ===== Hilfsfunktionen =====

// navigateTo baut eine Seite und zeigt sie an
func navigateTo(name string, builder func() tview.Primitive) {
	pages.RemovePage(name)
	pages.AddAndSwitchToPage(name, builder(), true)
}

// goBack kehrt zum Hauptmenü zurück
func goBack() {
	pages.SwitchToPage("main")
}

// addEscapeHandler fügt Escape-Handler zu einem Form hinzu
func addFormEscape(form *tview.Form, backFn func()) {
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backFn()
			return nil
		}
		return event
	})
}

// Verbindungsinfo für Header
func connInfo() string {
	if connMode == ConnSSH && sshConn != nil {
		return fmt.Sprintf("[#fe8019]SSH: %s@%s:%s[#665c54] → %s",
			sshConn.User, sshConn.Host, sshConn.Port, sshConn.RemotePath)
	}
	return fmt.Sprintf("[#665c54]Lokal: %s", docRoot)
}

// ===== Dialoge (custom, mit korrektem Button-Styling) =====

func showConfirm(title, message string, onConfirm func()) {
	form := styledForm()
	form.SetTitle(fmt.Sprintf(" %s ", title))

	// Nachricht als Label
	form.AddTextView("", message, 50, 2, true, false)

	form.AddButton("Nein", func() {
		pages.RemovePage("confirm")
	})
	form.AddButton("Ja", func() {
		pages.RemovePage("confirm")
		onConfirm()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.RemovePage("confirm")
			return nil
		}
		return event
	})

	// Zentriert anzeigen
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 10, 0, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(colorBg)

	pages.AddAndSwitchToPage("confirm", flex, true)
}

func showMessage(title, message string) {
	form := styledForm()
	form.SetTitle(fmt.Sprintf(" %s ", title))

	form.AddTextView("", message, 50, 2, true, false)

	form.AddButton("OK", func() {
		pages.RemovePage("message")
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
			pages.RemovePage("message")
			return nil
		}
		return event
	})

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 9, 0, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(colorBg)

	pages.AddAndSwitchToPage("message", flex, true)
}

// ===== Hauptmenü =====

func buildMainMenu() tview.Primitive {
	menu := styledList()
	menu.SetTitle(" ☰ Site Manager ")

	// Verbindungsstatus anzeigen
	wartungStatus := "[#b8bb26]● Online"
	if config.Wartung {
		wartungStatus = "[#fb4934]● Wartung aktiv"
	}

	menu.AddItem("⚙  Konfiguration", "config.json – Name, Kontakt, Bio, Terminal", 'k', func() {
		navigateTo("config", buildConfigEditor)
	})
	menu.AddItem("⚠  Wartungsmodus", fmt.Sprintf("Seite aktivieren/deaktivieren  %s", wartungStatus), 'w', func() {
		navigateTo("wartung", buildWartungEditor)
	})
	menu.AddItem("✎  Blog-Posts", fmt.Sprintf("Beiträge verwalten (%d Posts)", len(blog.Posts)), 'b', func() {
		navigateTo("blog-list", buildBlogList)
	})
	menu.AddItem("▣  Kategorien", fmt.Sprintf("Blog-Kategorien (%d)", len(blog.Categories)), 'g', func() {
		navigateTo("category-list", buildCategoryList)
	})
	menu.AddItem("§  Impressum", "impressum.json bearbeiten", 'i', func() {
		navigateTo("impressum", func() tview.Primitive {
			return buildLegalList("impressum", &impressum, "impressum.json")
		})
	})
	menu.AddItem("☷  Datenschutz", "datenschutz.json bearbeiten", 'd', func() {
		navigateTo("datenschutz", func() tview.Primitive {
			return buildLegalList("datenschutz", &datenschutz, "datenschutz.json")
		})
	})
	menu.AddItem("📂  Document Root", fmt.Sprintf("Aktuell: %s", docRoot), 'p', func() {
		navigateTo("docroot", buildDocRootChanger)
	})
	menu.AddItem("🔗  SSH/SFTP Verbindung", "Remote-Server verbinden", 's', func() {
		navigateTo("ssh", buildSSHForm)
	})
	menu.AddItem("↺  Neu laden", "Alle JSON-Dateien neu einlesen", 'r', func() {
		loadAll()
		navigateTo("main", buildMainMenu)
		showMessage("Neu geladen", "Alle Dateien wurden neu eingelesen.")
	})
	menu.AddItem("✕  Beenden", "Programm beenden", 'q', func() {
		if connMode == ConnSSH && sshConn != nil {
			sshConn.Close()
		}
		app.Stop()
	})

	// Header mit Verbindungsinfo
	header := styledTextView()
	header.SetTextAlign(tview.AlignCenter)
	header.SetBorder(false)
	header.SetText(fmt.Sprintf("[#b8bb26]Site Manager[#a89984] v3.0\n%s", connInfo()))

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(menu, 0, 1, true).
		AddItem(statusBar("↑↓: Navigation │ Enter/Buchstabe: Auswählen │ q: Beenden"), 1, 0, false)

	return layout
}

// ===== Document Root Wechsel =====

func buildDocRootChanger() tview.Primitive {
	form := styledForm()
	form.SetTitle(" 📂 Document Root ändern ")

	currentPath := docRoot
	if connMode == ConnSSH && sshConn != nil {
		currentPath = sshConn.RemotePath
	}

	form.AddInputField("Neuer Pfad", currentPath, 60, nil, nil)

	form.AddButton("Zurück", goBack)

	form.AddButton("Übernehmen", func() {
		newPath := form.GetFormItemByLabel("Neuer Pfad").(*tview.InputField).GetText()
		newPath = strings.TrimSpace(newPath)

		if newPath == "" {
			showMessage("Fehler", "Bitte einen Pfad eingeben.")
			return
		}

		if connMode == ConnSSH && sshConn != nil {
			// Remote: prüfe ob Verzeichnis existiert
			isDir, err := sshConn.StatDir(newPath)
			if err != nil || !isDir {
				showMessage("Fehler", fmt.Sprintf("Remote-Pfad ungültig: %s", newPath))
				return
			}
			sshConn.RemotePath = newPath
			loadAll()
			navigateTo("main", buildMainMenu)
			showMessage("Document Root", fmt.Sprintf("Remote-Pfad geändert: %s", newPath))
		} else {
			// Lokal: prüfe ob Verzeichnis existiert
			absPath, err := filepath.Abs(newPath)
			if err != nil {
				showMessage("Fehler", fmt.Sprintf("Ungültiger Pfad: %s", err))
				return
			}
			info, err := os.Stat(absPath)
			if err != nil || !info.IsDir() {
				showMessage("Fehler", fmt.Sprintf("%s ist kein gültiges Verzeichnis", absPath))
				return
			}
			docRoot = absPath
			loadAll()
			navigateTo("main", buildMainMenu)
			showMessage("Document Root", fmt.Sprintf("Pfad geändert: %s", absPath))
		}
	})
	addFormEscape(form, goBack)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 9, 0, true).
		AddItem(statusBar("Tab: Navigation │ Enter: Bestätigen │ Esc: Zurück"), 1, 0, false)

	return layout
}

// ===== SSH/SFTP Verbindung =====

func buildSSHForm() tview.Primitive {
	form := styledForm()
	form.SetTitle(" 🔗 SSH/SFTP Verbindung ")

	// Status-Anzeige
	statusText := styledTextView()
	statusText.SetTitle(" Status ")
	if connMode == ConnSSH && sshConn != nil {
		statusText.SetText(fmt.Sprintf("[#b8bb26]● Verbunden mit %s@%s:%s\n[#a89984]Remote-Pfad: %s",
			sshConn.User, sshConn.Host, sshConn.Port, sshConn.RemotePath))
	} else {
		statusText.SetText("[#a89984]Keine SSH-Verbindung aktiv (lokaler Modus)")
	}

	// Vorausgefüllte Werte
	defaultHost := ""
	defaultPort := "22"
	defaultUser := ""
	defaultPath := "/var/www/html"
	defaultKeyPath := ""

	home, _ := os.UserHomeDir()
	if home != "" {
		defaultKeyPath = filepath.Join(home, ".ssh", "id_rsa")
	}

	if sshConn != nil {
		if sshConn.Host != "" {
			defaultHost = sshConn.Host
		}
		if sshConn.Port != "" {
			defaultPort = sshConn.Port
		}
		if sshConn.User != "" {
			defaultUser = sshConn.User
		}
		if sshConn.RemotePath != "" {
			defaultPath = sshConn.RemotePath
		}
		if sshConn.KeyPath != "" {
			defaultKeyPath = sshConn.KeyPath
		}
	}

	form.AddInputField("Host", defaultHost, 40, nil, nil)
	form.AddInputField("Port", defaultPort, 8, nil, nil)
	form.AddInputField("Benutzer", defaultUser, 30, nil, nil)
	form.AddDropDown("Authentifizierung", []string{"SSH-Key", "Passwort"}, 0, nil)
	form.AddInputField("Key-Pfad", defaultKeyPath, 50, nil, nil)
	form.AddPasswordField("Passwort/Passphrase", "", 40, '*', nil)
	form.AddInputField("Remote-Pfad", defaultPath, 50, nil, nil)

	form.AddButton("Zurück", func() {
		navigateTo("main", buildMainMenu)
	})

	form.AddButton("Verbinden", func() {
		host := form.GetFormItemByLabel("Host").(*tview.InputField).GetText()
		port := form.GetFormItemByLabel("Port").(*tview.InputField).GetText()
		user := form.GetFormItemByLabel("Benutzer").(*tview.InputField).GetText()
		_, authOpt := form.GetFormItemByLabel("Authentifizierung").(*tview.DropDown).GetCurrentOption()
		keyPath := form.GetFormItemByLabel("Key-Pfad").(*tview.InputField).GetText()
		password := form.GetFormItemByLabel("Passwort/Passphrase").(*tview.InputField).GetText()
		remotePath := form.GetFormItemByLabel("Remote-Pfad").(*tview.InputField).GetText()

		if host == "" || user == "" {
			showMessage("Fehler", "Host und Benutzer sind Pflichtfelder.")
			return
		}
		if port == "" {
			port = "22"
		}

		// Alte Verbindung schließen
		if sshConn != nil {
			sshConn.Close()
		}

		authMethod := "key"
		if authOpt == "Passwort" {
			authMethod = "password"
		}

		sshConn = &SSHConnection{
			Host:       host,
			Port:       port,
			User:       user,
			AuthMethod: authMethod,
			Password:   password,
			KeyPath:    keyPath,
			RemotePath: remotePath,
		}

		statusText.SetText("[#fabd2f]⏳ Verbinde...")
		app.ForceDraw()

		if err := sshConn.Connect(); err != nil {
			statusText.SetText(fmt.Sprintf("[#fb4934]✕ Verbindung fehlgeschlagen:\n%s", err.Error()))
			sshConn = nil
			return
		}

		// Prüfe ob Remote-Pfad existiert
		isDir, err := sshConn.StatDir(remotePath)
		if err != nil || !isDir {
			statusText.SetText(fmt.Sprintf("[#fb4934]✕ Remote-Pfad nicht gefunden: %s", remotePath))
			sshConn.Close()
			sshConn = nil
			return
		}

		connMode = ConnSSH
		loadAll()
		statusText.SetText(fmt.Sprintf("[#b8bb26]● Verbunden mit %s@%s:%s\n[#a89984]Remote-Pfad: %s\n[#b8bb26]Dateien geladen!",
			user, host, port, remotePath))
	})

	if connMode == ConnSSH {
		form.AddButton("Trennen", func() {
			if sshConn != nil {
				sshConn.Close()
			}
			sshConn = nil
			connMode = ConnLocal
			loadAll()
			statusText.SetText("[#a89984]SSH-Verbindung getrennt. Lokaler Modus aktiv.")
		})
	}

	addFormEscape(form, func() {
		navigateTo("main", buildMainMenu)
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(statusText, 5, 0, false).
		AddItem(statusBar("Tab: Nächstes Feld │ Enter: Verbinden │ Esc: Zurück"), 1, 0, false)

	return layout
}

// ===== Konfiguration Editor =====

func buildConfigEditor() tview.Primitive {
	form := styledForm()
	form.SetTitle(" ⚙ Konfiguration ")

	form.AddInputField("Name", config.Name, 40, nil, func(text string) { config.Name = text })
	form.AddInputField("Initialen", config.Initials, 10, nil, func(text string) { config.Initials = text })
	form.AddInputField("Seitentitel", config.Seitentitel, 50, nil, func(text string) { config.Seitentitel = text })
	form.AddInputField("Jobtitel", config.JobTitle, 40, nil, func(text string) { config.JobTitle = text })
	form.AddInputField("Branche", config.Branche, 40, nil, func(text string) { config.Branche = text })
	form.AddInputField("Standort", config.Standort, 40, nil, func(text string) { config.Standort = text })
	form.AddInputField("Status", config.Status, 20, nil, func(text string) { config.Status = text })

	form.AddInputField("E-Mail", config.Kontakt.Email, 40, nil, func(text string) { config.Kontakt.Email = text })
	form.AddInputField("Telefon", config.Kontakt.Telefon, 30, nil, func(text string) { config.Kontakt.Telefon = text })
	form.AddInputField("Telefon Anzeige", config.Kontakt.TelefonAnzeige, 30, nil, func(text string) { config.Kontakt.TelefonAnzeige = text })

	form.AddInputField("Terminal User", config.Terminal.User, 20, nil, func(text string) { config.Terminal.User = text })
	form.AddInputField("Terminal Host", config.Terminal.Host, 20, nil, func(text string) { config.Terminal.Host = text })

	form.AddInputField("Skills (Komma-getrennt)", strings.Join(config.Skills, ", "), 60, nil, func(text string) {
		parts := strings.Split(text, ",")
		config.Skills = make([]string, 0)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				config.Skills = append(config.Skills, p)
			}
		}
	})

	form.AddTextArea("Bio", config.Bio, 60, 4, 0, func(text string) { config.Bio = text })

	// Impressum-Daten
	form.AddInputField("Impressum Name", config.Impressum.Name, 40, nil, func(text string) { config.Impressum.Name = text })
	form.AddInputField("Impressum Straße", config.Impressum.Strasse, 40, nil, func(text string) { config.Impressum.Strasse = text })
	form.AddInputField("Impressum PLZ/Ort", config.Impressum.PlzOrt, 40, nil, func(text string) { config.Impressum.PlzOrt = text })
	form.AddInputField("Impressum Land", config.Impressum.Land, 40, nil, func(text string) { config.Impressum.Land = text })
	form.AddInputField("Impressum Telefon", config.Impressum.Telefon, 30, nil, func(text string) { config.Impressum.Telefon = text })
	form.AddInputField("Impressum E-Mail", config.Impressum.Email, 40, nil, func(text string) { config.Impressum.Email = text })

	form.AddButton("Zurück", func() {
		navigateTo("main", buildMainMenu)
	})
	form.AddButton("Speichern", func() {
		if err := saveJSON("config.json", config); err != nil {
			showMessage("Fehler", "Speichern fehlgeschlagen: "+err.Error())
		} else {
			showMessage("Gespeichert", "config.json wurde gespeichert.")
		}
	})

	addFormEscape(form, func() {
		navigateTo("main", buildMainMenu)
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(statusBar("Tab: Nächstes Feld │ Shift+Tab: Vorheriges │ Esc: Zurück"), 1, 0, false)

	return layout
}

// ===== Wartungsmodus =====

func buildWartungEditor() tview.Primitive {
	form := styledForm()
	form.SetTitle(" ⚠ Wartungsmodus ")

	form.AddCheckbox("Wartungsmodus aktiv", config.Wartung, func(checked bool) {
		config.Wartung = checked
	})

	form.AddInputField("Wartungstext", config.WartungText, 60, nil, func(text string) {
		config.WartungText = text
	})

	form.AddButton("Zurück", func() {
		navigateTo("main", buildMainMenu)
	})
	form.AddButton("Speichern", func() {
		if err := saveJSON("config.json", config); err != nil {
			showMessage("Fehler", "Speichern fehlgeschlagen: "+err.Error())
		} else {
			statusText := "DEAKTIVIERT"
			if config.Wartung {
				statusText = "AKTIVIERT"
			}
			showMessage("Gespeichert", fmt.Sprintf("Wartungsmodus: %s", statusText))
		}
	})

	addFormEscape(form, func() {
		navigateTo("main", buildMainMenu)
	})

	// Vorschau – wird bei jedem Öffnen neu erstellt
	preview := styledTextView()
	preview.SetTitle(" Vorschau ")
	if config.Wartung {
		preview.SetText(fmt.Sprintf("[#fabd2f::b]⚠ WARTUNG AKTIV\n\n[#a89984]%s", config.WartungText))
	} else {
		preview.SetText("[#b8bb26]✓ Website ist online")
	}

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 10, 0, true).
		AddItem(preview, 5, 0, false).
		AddItem(statusBar("Tab: Navigation │ Leertaste: Checkbox │ Esc: Zurück"), 1, 0, false)

	return flex
}

// ===== Blog Post Liste =====

func buildBlogList() tview.Primitive {
	list := styledList()
	list.SetTitle(fmt.Sprintf(" ✎ Blog-Posts (%d) ", len(blog.Posts)))

	for i, post := range blog.Posts {
		cats := post.GetCategories()
		catStr := strings.Join(cats, ", ")
		idx := i
		list.AddItem(
			fmt.Sprintf("%s  [%s]", post.Title, post.Date),
			fmt.Sprintf("Kategorien: %s │ ID: %s", catStr, post.ID),
			0,
			func() {
				navigateTo("blog-edit", func() tview.Primitive {
					return buildBlogEditor(idx)
				})
			},
		)
	}

	list.AddItem("[#b8bb26]＋ Neuen Post erstellen", "", 'n', func() {
		newPost := BlogPost{
			ID:       "neuer-post",
			Title:    "Neuer Blogpost",
			Date:     time.Now().Format("02-01-2006"),
			Category: "alltag",
			Tags:     []string{},
			Excerpt:  "",
			Content:  "## Überschrift\n\nInhalt hier...",
		}
		blog.Posts = append([]BlogPost{newPost}, blog.Posts...)
		navigateTo("blog-edit", func() tview.Primitive {
			return buildBlogEditor(0)
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			navigateTo("main", buildMainMenu)
			return nil
		}
		if event.Key() == tcell.KeyDelete || event.Rune() == 'x' {
			idx := list.GetCurrentItem()
			if idx >= 0 && idx < len(blog.Posts) {
				postTitle := blog.Posts[idx].Title
				showConfirm("Löschen", fmt.Sprintf("Post \"%s\" wirklich löschen?", postTitle), func() {
					blog.Posts = append(blog.Posts[:idx], blog.Posts[idx+1:]...)
					if err := saveJSON("blog.json", blog); err != nil {
						showMessage("Fehler", err.Error())
					}
					navigateTo("blog-list", buildBlogList)
				})
			}
			return nil
		}
		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(statusBar("Enter: Bearbeiten │ n: Neuer Post │ x: Löschen │ Esc: Zurück"), 1, 0, false)

	return layout
}

// ===== Blog Post Editor =====

func buildBlogEditor(index int) tview.Primitive {
	post := &blog.Posts[index]

	form := styledForm()
	form.SetTitle(fmt.Sprintf(" ✎ Post: %s ", post.Title))

	form.AddInputField("ID (URL-Slug)", post.ID, 40, nil, func(text string) { post.ID = text })
	form.AddInputField("Titel", post.Title, 60, nil, func(text string) { post.Title = text })
	form.AddInputField("Datum", post.Date, 20, nil, func(text string) { post.Date = text })

	form.AddInputField("Tags (Komma-getrennt)", strings.Join(post.Tags, ", "), 60, nil, func(text string) {
		parts := strings.Split(text, ",")
		post.Tags = make([]string, 0)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				post.Tags = append(post.Tags, p)
			}
		}
	})

	form.AddInputField("Kurztext (Excerpt)", post.Excerpt, 60, nil, func(text string) { post.Excerpt = text })
	form.AddTextArea("Inhalt (Markdown)", post.Content, 70, 10, 0, func(text string) { post.Content = text })

	// Kategorien als kompaktes Grid (Table)
	currentCats := post.GetCategories()
	catCheckStates := make(map[string]bool)
	for _, c := range currentCats {
		catCheckStates[c] = true
	}

	catTable := tview.NewTable()
	catTable.SetBackgroundColor(colorBg)
	catTable.SetBorder(true)
	catTable.SetBorderColor(colorBgCard)
	catTable.SetTitle(" Kategorien (Enter/Leertaste: umschalten) ")
	catTable.SetTitleColor(colorGreen)
	catTable.SetSelectable(true, true)

	cols := 4 // Spalten im Grid
	updateCatTable := func() {
		catTable.Clear()
		for i, cat := range blog.Categories {
			row := i / cols
			col := i % cols
			marker := "[ ]"
			style := tcell.StyleDefault.Background(colorBg).Foreground(colorTextDim)
			if catCheckStates[cat.ID] {
				marker = "[x]"
				style = tcell.StyleDefault.Background(colorBg).Foreground(colorGreen)
			}
			cell := tview.NewTableCell(fmt.Sprintf(" %s %s ", marker, cat.Name)).
				SetStyle(style).
				SetAlign(tview.AlignLeft).
				SetExpansion(1)
			catTable.SetCell(row, col, cell)
		}
	}
	updateCatTable()

	syncCatsToPost := func() {
		newCats := make([]string, 0)
		for _, bc := range blog.Categories {
			if catCheckStates[bc.ID] {
				newCats = append(newCats, bc.ID)
			}
		}
		post.SetCategories(newCats)
	}

	catTable.SetSelectedFunc(func(row, col int) {
		idx := row*cols + col
		if idx < len(blog.Categories) {
			catID := blog.Categories[idx].ID
			catCheckStates[catID] = !catCheckStates[catID]
			syncCatsToPost()
			updateCatTable()
		}
	})

	// Buttons
	blogBackFn := func() {
		navigateTo("blog-list", buildBlogList)
	}

	form.AddButton("Zurück", blogBackFn)
	form.AddButton("Speichern", func() {
		if err := saveJSON("blog.json", blog); err != nil {
			showMessage("Fehler", "Speichern fehlgeschlagen: "+err.Error())
		} else {
			showMessage("Gespeichert", fmt.Sprintf("Post \"%s\" wurde gespeichert.", post.Title))
		}
	})

	// Tab-Cycling: Form → Kategorien-Grid → Form
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			blogBackFn()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			_, btnIdx := form.GetFocusedItemIndex()
			if btnIdx == form.GetButtonCount()-1 {
				app.SetFocus(catTable)
				return nil
			}
		}
		return event
	})

	catTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == ' ' {
			row, col := catTable.GetSelection()
			idx := row*cols + col
			if idx < len(blog.Categories) {
				catID := blog.Categories[idx].ID
				catCheckStates[catID] = !catCheckStates[catID]
				syncCatsToPost()
				updateCatTable()
			}
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			blogBackFn()
			return nil
		}
		// Tab oder Backtab → zurück zum Form
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
			app.SetFocus(form)
			return nil
		}
		return event
	})

	// Höhe des Kategorie-Grids berechnen
	catRows := (len(blog.Categories) + cols - 1) / cols
	catHeight := catRows + 2 // +2 für Border

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(catTable, catHeight, 0, false).
		AddItem(statusBar("Tab: Formular ↔ Kategorien │ Leertaste/Enter: umschalten │ Esc: Zurück"), 1, 0, false)

	return layout
}

// ===== Kategorie-Manager =====

func buildCategoryList() tview.Primitive {
	list := styledList()
	list.SetTitle(fmt.Sprintf(" ▣ Kategorien (%d) ", len(blog.Categories)))

	for i, cat := range blog.Categories {
		idx := i
		list.AddItem(
			cat.Name,
			fmt.Sprintf("ID: %s │ Farbe: %s", cat.ID, cat.Color),
			0,
			func() {
				navigateTo("category-edit", func() tview.Primitive {
					return buildCategoryEditor(idx)
				})
			},
		)
	}

	list.AddItem("[#b8bb26]＋ Neue Kategorie", "", 'n', func() {
		blog.Categories = append(blog.Categories, BlogCategory{
			ID:    "neue-kategorie",
			Name:  "Neue Kategorie",
			Color: "#b8bb26",
		})
		navigateTo("category-edit", func() tview.Primitive {
			return buildCategoryEditor(len(blog.Categories) - 1)
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			navigateTo("main", buildMainMenu)
			return nil
		}
		if event.Key() == tcell.KeyDelete || event.Rune() == 'x' {
			idx := list.GetCurrentItem()
			if idx >= 0 && idx < len(blog.Categories) {
				catName := blog.Categories[idx].Name
				showConfirm("Löschen", fmt.Sprintf("Kategorie \"%s\" wirklich löschen?", catName), func() {
					blog.Categories = append(blog.Categories[:idx], blog.Categories[idx+1:]...)
					if err := saveJSON("blog.json", blog); err != nil {
						showMessage("Fehler", err.Error())
					}
					navigateTo("category-list", buildCategoryList)
				})
			}
			return nil
		}
		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(statusBar("Enter: Bearbeiten │ n: Neue Kategorie │ x: Löschen │ Esc: Zurück"), 1, 0, false)

	return layout
}

// Farbpalette – 20 Farbtöne (Gruvbox-kompatibel + weitere)
var colorPalette = []struct {
	Hex  string
	Name string
}{
	{"#fb4934", "Rot"},
	{"#cc241d", "Dunkelrot"},
	{"#fe8019", "Orange"},
	{"#d65d0e", "Dunkelorange"},
	{"#fabd2f", "Gelb"},
	{"#d79921", "Dunkelgelb"},
	{"#b8bb26", "Grün"},
	{"#98971a", "Dunkelgrün"},
	{"#8ec07c", "Mintgrün"},
	{"#689d6a", "Waldgrün"},
	{"#83a598", "Blaugrün"},
	{"#458588", "Petrol"},
	{"#7daea3", "Hellblau"},
	{"#076678", "Dunkelblau"},
	{"#d3869b", "Rosa"},
	{"#b16286", "Magenta"},
	{"#d4879c", "Altrosa"},
	{"#a9b665", "Limette"},
	{"#e78a4e", "Lachs"},
	{"#ea6962", "Koralle"},
}

func buildCategoryEditor(index int) tview.Primitive {
	cat := &blog.Categories[index]

	form := styledForm()
	form.SetTitle(fmt.Sprintf(" ▣ Kategorie: %s ", cat.Name))

	form.AddInputField("ID", cat.ID, 30, nil, func(text string) { cat.ID = text })
	form.AddInputField("Name", cat.Name, 40, nil, func(text string) { cat.Name = text })

	// Farbe als Eingabefeld (manuell überschreibbar)
	var colorField *tview.InputField
	form.AddInputField("Farbe (Hex)", cat.Color, 10, nil, func(text string) { cat.Color = text })
	colorField = form.GetFormItemByLabel("Farbe (Hex)").(*tview.InputField)

	form.AddButton("Zurück", func() {
		navigateTo("category-list", buildCategoryList)
	})
	form.AddButton("Speichern", func() {
		if err := saveJSON("blog.json", blog); err != nil {
			showMessage("Fehler", err.Error())
		} else {
			showMessage("Gespeichert", fmt.Sprintf("Kategorie \"%s\" gespeichert.", cat.Name))
		}
	})

	// Farbvorschau
	previewText := styledTextView()
	previewText.SetTitle(" Vorschau ")
	previewText.SetBorder(true)
	updatePreview := func() {
		previewText.SetText(fmt.Sprintf("[%s]████████████  %s  ████████████", cat.Color, cat.Name))
	}
	updatePreview()

	// Farbpalette als Table (5 Spalten × 4 Reihen = 20 Farben)
	palTable := tview.NewTable()
	palTable.SetBackgroundColor(colorBg)
	palTable.SetBorder(true)
	palTable.SetBorderColor(colorBgCard)
	palTable.SetTitle(" Farbauswahl (Enter: übernehmen) ")
	palTable.SetTitleColor(colorGreen)
	palTable.SetSelectable(true, true)

	palCols := 5
	buildPalCells := func() {
		palTable.Clear()
		for i, pc := range colorPalette {
			row := i / palCols
			col := i % palCols
			marker := " "
			if strings.EqualFold(pc.Hex, cat.Color) {
				marker = "●"
			}
			cell := tview.NewTableCell(fmt.Sprintf(" %s ████ %s ", marker, pc.Name)).
				SetStyle(tcell.StyleDefault.Background(colorBg).Foreground(tcell.NewRGBColor(
					hexToRGB(pc.Hex)))).
				SetAlign(tview.AlignLeft).
				SetExpansion(1)
			palTable.SetCell(row, col, cell)
		}
	}
	buildPalCells()

	palTable.SetSelectedFunc(func(row, col int) {
		idx := row*palCols + col
		if idx < len(colorPalette) {
			cat.Color = colorPalette[idx].Hex
			colorField.SetText(cat.Color)
			updatePreview()
			buildPalCells()
		}
	})

	// Tab-Cycling: Form → Farbpalette → Form
	backFn := func() {
		navigateTo("category-list", buildCategoryList)
	}

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backFn()
			return nil
		}
		// Tab auf letztem Button → Fokus zur Farbpalette
		if event.Key() == tcell.KeyTab {
			_, btnIdx := form.GetFocusedItemIndex()
			if btnIdx == form.GetButtonCount()-1 {
				app.SetFocus(palTable)
				return nil
			}
		}
		return event
	})

	palTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backFn()
			return nil
		}
		// Shift+Tab oder Backtab → zurück zum Form
		if event.Key() == tcell.KeyBacktab {
			app.SetFocus(form)
			return nil
		}
		// Tab am Ende der Palette → zurück zum Form
		if event.Key() == tcell.KeyTab {
			app.SetFocus(form)
			return nil
		}
		return event
	})

	palHeight := (len(colorPalette)+palCols-1)/palCols + 2

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 10, 0, true).
		AddItem(palTable, palHeight, 0, false).
		AddItem(previewText, 3, 0, false).
		AddItem(statusBar("Tab: Formular ↔ Farbpalette │ Enter: Farbe/Speichern │ Esc: Zurück"), 1, 0, false)

	return layout
}

// hexToRGB konvertiert "#rrggbb" zu int32 r, g, b
func hexToRGB(hex string) (int32, int32, int32) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 180, 180, 180
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return int32(r), int32(g), int32(b)
}

// ===== Impressum / Datenschutz Editor =====

func buildLegalList(pageName string, data *LegalData, filename string) tview.Primitive {
	list := styledList()
	list.SetTitle(fmt.Sprintf(" § %s (%d Abschnitte) ", data.Titel, len(data.Abschnitte)))

	for i, section := range data.Abschnitte {
		idx := i
		prefix := ""
		if section.Nr != "" {
			prefix = section.Nr + ". "
		}
		preview := section.Inhalt
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}
		preview = strings.ReplaceAll(preview, "\n", " ")

		list.AddItem(
			prefix+section.Ueberschrift,
			preview,
			0,
			func() {
				navigateTo(pageName+"-edit", func() tview.Primitive {
					return buildLegalSectionEditor(pageName, data, filename, idx)
				})
			},
		)
	}

	list.AddItem("[#b8bb26]＋ Neuen Abschnitt hinzufügen", "", 'n', func() {
		newNr := ""
		if len(data.Abschnitte) > 0 {
			lastNr := data.Abschnitte[len(data.Abschnitte)-1].Nr
			if lastNr != "" {
				if n, err := strconv.Atoi(lastNr); err == nil {
					newNr = strconv.Itoa(n + 1)
				}
			}
		}
		data.Abschnitte = append(data.Abschnitte, LegalSection{
			Nr:           newNr,
			Ueberschrift: "Neuer Abschnitt",
			Inhalt:       "",
		})
		navigateTo(pageName+"-edit", func() tview.Primitive {
			return buildLegalSectionEditor(pageName, data, filename, len(data.Abschnitte)-1)
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			navigateTo("main", buildMainMenu)
			return nil
		}
		if event.Key() == tcell.KeyDelete || event.Rune() == 'x' {
			idx := list.GetCurrentItem()
			if idx >= 0 && idx < len(data.Abschnitte) {
				sectionName := data.Abschnitte[idx].Ueberschrift
				showConfirm("Löschen", fmt.Sprintf("Abschnitt \"%s\" wirklich löschen?", sectionName), func() {
					data.Abschnitte = append(data.Abschnitte[:idx], data.Abschnitte[idx+1:]...)
					if err := saveJSON(filename, data); err != nil {
						showMessage("Fehler", err.Error())
					}
					navigateTo(pageName, func() tview.Primitive {
						return buildLegalList(pageName, data, filename)
					})
				})
			}
			return nil
		}
		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(statusBar("Enter: Bearbeiten │ n: Neuer Abschnitt │ x: Löschen │ Esc: Zurück"), 1, 0, false)

	return layout
}

func buildLegalSectionEditor(pageName string, data *LegalData, filename string, index int) tview.Primitive {
	section := &data.Abschnitte[index]

	form := styledForm()
	form.SetTitle(fmt.Sprintf(" § %s ", section.Ueberschrift))

	if section.Nr != "" || pageName == "datenschutz" {
		form.AddInputField("Nummer", section.Nr, 5, nil, func(text string) { section.Nr = text })
	}
	form.AddInputField("Überschrift", section.Ueberschrift, 60, nil, func(text string) { section.Ueberschrift = text })
	form.AddTextArea("Inhalt", section.Inhalt, 70, 12, 0, func(text string) { section.Inhalt = text })

	backFn := func() {
		navigateTo(pageName, func() tview.Primitive {
			return buildLegalList(pageName, data, filename)
		})
	}
	form.AddButton("Zurück", backFn)
	form.AddButton("Speichern", func() {
		if err := saveJSON(filename, data); err != nil {
			showMessage("Fehler", err.Error())
		} else {
			showMessage("Gespeichert", fmt.Sprintf("Abschnitt \"%s\" gespeichert.", section.Ueberschrift))
		}
	})
	addFormEscape(form, backFn)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(statusBar("Tab: Navigation │ Enter: Speichern │ Esc: Zurück │ {{impressum.*}} Platzhalter werden automatisch ersetzt"), 1, 0, false)

	return layout
}

// ===== MAIN =====

func main() {
	// CLI Flags
	flag.StringVar(&docRoot, "root", "/var/www/html", "Document Root – Pfad zum Verzeichnis mit den JSON-Dateien")
	flag.StringVar(&docRoot, "r", "/var/www/html", "Document Root (Kurzform)")
	flag.Parse()

	// Auch erstes Argument ohne Flag als Root akzeptieren
	if flag.NArg() > 0 && docRoot == "/var/www/html" {
		docRoot = flag.Arg(0)
	}

	// Pfad auflösen
	absRoot, err := filepath.Abs(docRoot)
	if err == nil {
		docRoot = absRoot
	}

	// Prüfen ob Verzeichnis existiert – wenn nicht, trotzdem starten
	info, statErr := os.Stat(docRoot)
	if statErr != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Hinweis: %s existiert nicht oder ist kein Verzeichnis.\n", docRoot)
		fmt.Fprintf(os.Stderr, "Das Programm startet trotzdem – bitte Document Root im Menü anpassen.\n")
	}

	// Initialer Modus: lokal
	connMode = ConnLocal

	// JSON-Dateien laden
	loadAll()

	// TUI aufbauen
	app = tview.NewApplication()
	pages = tview.NewPages()

	pages.AddPage("main", buildMainMenu(), true, true)

	// Globale Escape-Taste → im Hauptmenü beendet die App
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		name, _ := pages.GetFrontPage()
		if event.Key() == tcell.KeyEscape && name == "main" {
			if connMode == ConnSSH && sshConn != nil {
				sshConn.Close()
			}
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %s\n", err)
		os.Exit(1)
	}
}
