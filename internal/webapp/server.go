package webapp

import (
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"

	"perezvonish/factorio-server-manager/internal/factorio/saves"
)

//go:embed static/upload.html
var uploadHTML []byte

// Server is a lightweight HTTP server that serves the save-upload WebApp.
type Server struct {
	botToken     string
	allowedUsers map[int64]struct{}
	saves        *saves.Manager
	ready        atomic.Bool // true after initial SyncMods completes
}

func NewServer(botToken string, allowedUsers map[int64]struct{}, saves *saves.Manager) *Server {
	return &Server{
		botToken:     botToken,
		allowedUsers: allowedUsers,
		saves:        saves,
	}
}

// SetReady signals that startup (SyncMods) has completed.
// After this, /health returns 200 and the Factorio container may start.
func (s *Server) SetReady() { s.ready.Store(true) }

func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/upload", s.handleUpload)
	log.Printf("webapp: listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

// handleHealth is used by the Docker health check.
// Returns 200 OK only after SetReady() has been called (i.e. SyncMods completed).
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	if s.ready.Load() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("starting")) //nolint:errcheck
	}
}

// handleIndex serves the upload HTML page.
func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(uploadHTML) //nolint:errcheck
}

// handleUpload accepts a multipart POST with a "save" file field.
// The request must carry a valid Telegram initData in X-Telegram-Init-Data header.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-Telegram-Init-Data")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ── auth ──────────────────────────────────────────────────────────────
	initData := r.Header.Get("X-Telegram-Init-Data")
	userID, ok := s.validateInitData(initData)
	if !ok {
		log.Printf("webapp: invalid initData from %s", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if _, allowed := s.allowedUsers[userID]; !allowed {
		log.Printf("webapp: user %d is not in allowed list", userID)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// ── parse file ────────────────────────────────────────────────────────
	if err := r.ParseMultipartForm(500 << 20); err != nil { // 500 MB max
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("save")
	if err != nil {
		http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		http.Error(w, "only .zip files are allowed", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.saves.Replace(header.Filename, data); err != nil {
		http.Error(w, "save error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("webapp: user %d uploaded save %q (%d bytes)", userID, header.Filename, len(data))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
		"status":   "ok",
		"filename": header.Filename,
	})
}

// ── Telegram WebApp initData validation ──────────────────────────────────────
//
// Algorithm: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
//  1. Parse initData as URL query string.
//  2. Extract the "hash" field.
//  3. Sort remaining key=value pairs alphabetically, join with "\n".
//  4. secret_key = HMAC-SHA256("WebAppData", bot_token)
//  5. computed   = HMAC-SHA256(data_check_string, secret_key)
//  6. Compare computed == hash (constant-time).

func (s *Server) validateInitData(initData string) (int64, bool) {
	if initData == "" {
		return 0, false
	}

	vals, err := url.ParseQuery(initData)
	if err != nil {
		return 0, false
	}

	hash := vals.Get("hash")
	if hash == "" {
		return 0, false
	}

	var parts []string
	for k, vs := range vals {
		if k == "hash" || len(vs) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, vs[0]))
	}
	sort.Strings(parts)
	dataCheckString := strings.Join(parts, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(s.botToken))
	secretKey := secretMAC.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedHash), []byte(hash)) {
		return 0, false
	}

	userJSON := vals.Get("user")
	if userJSON == "" {
		return 0, false
	}
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return 0, false
	}

	return user.ID, true
}
