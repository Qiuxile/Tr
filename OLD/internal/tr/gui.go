package tr

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

//go:embed gui/index.html
var guiHTML embed.FS

// StartGUI launches the local web GUI server and opens the browser.
func StartGUI(cfg Config, assetsFS embed.FS) error {
	port, err := findAvailablePort(8765)
	if err != nil {
		return fmt.Errorf("no available port: %w", err)
	}

	mux := http.NewServeMux()

	// Main page — inject config and i18n data
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveIndex(w, cfg)
	})

	// Static assets (icon)
	assetsSub, _ := fs.Sub(assetsFS, "assets")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsSub))))

	// API: translate
	mux.HandleFunc("/api/translate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleAPITranslate(w, r, cfg)
	})

	// API: get config
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleAPIGetConfig(w, cfg)
		case http.MethodPost:
			handleAPISetConfig(w, r, &cfg)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("GUI 服务已启动: %s\n", url)
	fmt.Println("按 Ctrl+C 退出")

	// Auto-open browser
	go openBrowser(url)

	return http.ListenAndServe(addr, mux)
}

// serveIndex reads the embedded HTML, injects config and i18n JSON, and serves it.
func serveIndex(w http.ResponseWriter, cfg Config) {
	data, err := guiHTML.ReadFile("gui/index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	configJSON, _ := json.Marshal(map[string]interface{}{
		"source_lang": cfg.SourceLang,
		"target_lang": cfg.TargetLang,
		"api_url":     cfg.ApiURL,
		"ui_lang":     cfg.UILang,
		"version":     Version,
	})

	// Build i18n JSON for zh, en, ja — only GUI-relevant keys
	guiKeys := []string{
		"app.name", "app.license",
		"gui.translate", "gui.input_placeholder", "gui.source_lang", "gui.target_lang",
		"gui.offline_mode", "gui.translate_btn", "gui.result", "gui.result_placeholder",
		"gui.settings", "gui.ui_lang", "gui.save", "gui.free", "gui.custom_api",
		"gui.ready", "gui.empty_input", "gui.translating", "gui.cache", "gui.dict",
		"gui.network_error", "gui.saved", "gui.save_failed", "gui.offline",
	}
	i18nJSON := buildI18nJSON(guiKeys)

	html := string(data)
	html = strings.Replace(html,
		"window.__I18N__ = window.__I18N__ || { zh: {}, en: {}, ja: {} };",
		"window.__I18N__ = "+i18nJSON+";", 1)
	html = strings.Replace(html,
		"window.__CONFIG__ = window.__CONFIG__ || {};",
		"window.__CONFIG__ = "+string(configJSON)+";", 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// buildI18nJSON creates a JSON object with only GUI-related keys for the given languages.
func buildI18nJSON(keys []string) string {
	langs := []string{"zh", "en", "ja"}
	result := make(map[string]map[string]string)
	for _, lang := range langs {
		result[lang] = make(map[string]string)
		for _, key := range keys {
			result[lang][key] = T(lang, key)
		}
	}
	data, _ := json.Marshal(result)
	return string(data)
}

// handleAPITranslate processes a translation request.
func handleAPITranslate(w http.ResponseWriter, r *http.Request, cfg Config) {
	var req struct {
		Text    string `json:"text"`
		Source  string `json:"source"`
		Target  string `json:"target"`
		Offline bool   `json:"offline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": "invalid request"})
		return
	}

	// Temporarily override source/target for this request
	reqCfg := cfg
	if req.Source != "" {
		reqCfg.SourceLang = req.Source
	}
	if req.Target != "" {
		reqCfg.TargetLang = req.Target
	}

	opts := Options{Offline: req.Offline}
	cache, _ := NewCache() // load cache fresh for each request

	result, err := Translate(reqCfg, opts, cache, req.Text)
	if err != nil {
		writeJSON(w, map[string]string{"error": T(cfg.UILang, "err.translate_failed", err)})
		return
	}
	writeJSON(w, map[string]string{
		"text":   result.Text,
		"source": result.Source,
	})
}

// handleAPIGetConfig returns the current config.
func handleAPIGetConfig(w http.ResponseWriter, cfg Config) {
	writeJSON(w, map[string]string{
		"source_lang": cfg.SourceLang,
		"target_lang": cfg.TargetLang,
		"api_url":     cfg.ApiURL,
		"ui_lang":     cfg.UILang,
	})
}

// handleAPISetConfig updates the config and persists it.
func handleAPISetConfig(w http.ResponseWriter, r *http.Request, cfg *Config) {
	var req struct {
		SourceLang string `json:"source_lang"`
		TargetLang string `json:"target_lang"`
		ApiURL     string `json:"api_url"`
		UILang     string `json:"ui_lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "invalid request"})
		return
	}
	if req.SourceLang != "" {
		cfg.SourceLang = req.SourceLang
	}
	if req.TargetLang != "" {
		cfg.TargetLang = req.TargetLang
	}
	if req.ApiURL != "" {
		cfg.ApiURL = req.ApiURL
	}
	if req.UILang != "" {
		cfg.UILang = req.UILang
	}
	if err := cfg.Save(); err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

// findAvailablePort tries ports starting from startPort.
func findAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port in range %d-%d", startPort, startPort+99)
}

// openBrowser opens the given URL in the system default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "无法自动打开浏览器，请手动访问: %s\n", url)
	}
}
