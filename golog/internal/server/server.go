// Package server is GoLog's HTTP API and serves the web app.
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"path"
	"strings"

	"golog/internal/model"
	"golog/internal/sedlog"
	"golog/internal/stats"
	"golog/internal/store"
)

// Version is set at build time.
var Version = "dev"

const maxBody = 8 << 20 // a log with 10,000 beds is well under this

// Server holds what the handlers need.
type Server struct {
	store *store.Store
	web   fs.FS  // the built web app (index.html, assets/)
	quit  func() // stops the program; nil if the app may not stop it
}

// New returns the HTTP handler for the API and the web app. quit, if not nil, is called when
// someone presses "Stop GoLog" in the app.
func New(st *store.Store, web fs.FS, quit func()) http.Handler {
	s := &Server{store: st, web: web, quit: quit}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("POST /api/quit", s.stop)
	mux.HandleFunc("GET /api/projects", s.listProjects)
	mux.HandleFunc("POST /api/projects", s.createProject)
	mux.HandleFunc("GET /api/projects/{id}", s.getProject)
	mux.HandleFunc("PATCH /api/projects/{id}", s.renameProject)
	mux.HandleFunc("DELETE /api/projects/{id}", s.deleteProject)
	mux.HandleFunc("POST /api/projects/{id}/logs", s.createLog)
	mux.HandleFunc("GET /api/logs/{id}", s.getLog)
	mux.HandleFunc("PUT /api/logs/{id}", s.updateLog)
	mux.HandleFunc("DELETE /api/logs/{id}", s.deleteLog)
	mux.HandleFunc("POST /api/import", s.importFile)
	mux.HandleFunc("POST /api/stats", s.statistics)
	mux.HandleFunc("/", s.static)
	return guard(mux)
}

// guard protects a server that has no login: only this computer may use it, and other
// websites open in the browser may not send it requests (DNS rebinding, cross-site POSTs).
func guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !localHost(r.Host) {
			http.Error(w, "GoLog only answers requests addressed to localhost", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
				fail(w, http.StatusForbidden, "request from another website refused")
				return
			}
			// A JSON content type cannot be sent cross-site without a CORS preflight, which we never allow.
			if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" && r.ContentLength != 0 {
				fail(w, http.StatusUnsupportedMediaType, "send JSON")
				return
			}
		}
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func localHost(hostport string) bool {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	reply(w, http.StatusOK, map[string]any{"app": "golog", "version": Version, "canStop": s.quit != nil})
}

// stop ends the program, for people who started it by double-clicking and have no terminal to
// press Ctrl+C in. Every change is saved as it is made, so nothing is lost.
func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	if s.quit == nil {
		fail(w, http.StatusNotFound, "GoLog can't be stopped from here")
		return
	}
	reply(w, http.StatusAccepted, map[string]string{"status": "stopping"})
	go s.quit()
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.Projects(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	reply(w, http.StatusOK, list)
}

type nameBody struct {
	Name string `json:"name"`
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var b nameBody
	if !decode(w, r, &b) {
		return
	}
	name, err := model.CleanName(b.Name)
	if err != nil {
		fail(w, http.StatusBadRequest, "project "+err.Error())
		return
	}
	p, err := s.store.CreateProject(r.Context(), name)
	if err != nil {
		internal(w, err)
		return
	}
	reply(w, http.StatusCreated, p)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.Project(r.Context(), r.PathValue("id"))
	if s.storeErr(w, err, "project") {
		return
	}
	reply(w, http.StatusOK, p)
}

func (s *Server) renameProject(w http.ResponseWriter, r *http.Request) {
	var b nameBody
	if !decode(w, r, &b) {
		return
	}
	name, err := model.CleanName(b.Name)
	if err != nil {
		fail(w, http.StatusBadRequest, "project "+err.Error())
		return
	}
	if s.storeErr(w, s.store.RenameProject(r.Context(), r.PathValue("id"), name), "project") {
		return
	}
	s.getProject(w, r)
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	if s.storeErr(w, s.store.DeleteProject(r.Context(), r.PathValue("id")), "project") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// createLog adds an empty log, or one with content (e.g. from an import).
func (s *Server) createLog(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Name    string            `json:"name"`
		Content *model.LogContent `json:"content"`
	}
	if !decode(w, r, &b) {
		return
	}
	name, err := model.CleanName(b.Name)
	if err != nil {
		fail(w, http.StatusBadRequest, "log "+err.Error())
		return
	}
	var content []model.LogContent
	if b.Content != nil {
		if err := b.Content.Validate(); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		content = append(content, *b.Content)
	}
	l, err := s.store.CreateLog(r.Context(), r.PathValue("id"), name, content...)
	if s.storeErr(w, err, "project") {
		return
	}
	reply(w, http.StatusCreated, l)
}

// importFile reads a SedLog file sent as base64 and returns it as GoLog content, without saving it.
func (s *Server) importFile(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Data []byte             `json:"data"` // base64 in JSON
		CSV  *sedlog.CSVOptions `json:"csv"`
	}
	if !decode(w, r, &b) {
		return
	}
	var res *sedlog.Result
	var err error
	if trimmed := bytes.TrimSpace(bytes.TrimPrefix(b.Data, []byte("\xef\xbb\xbf"))); bytes.HasPrefix(trimmed, []byte("{")) {
		res, err = readGoLogFile(trimmed)
	} else {
		res, err = sedlog.Parse(b.Data, b.CSV)
	}
	if err != nil {
		fail(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := res.Content.Validate(); err != nil {
		fail(w, http.StatusUnprocessableEntity, "the imported log is not valid: "+err.Error())
		return
	}
	reply(w, http.StatusOK, res)
}

// readGoLogFile reads a log saved with "Export → GoLog file".
func readGoLogFile(data []byte) (*sedlog.Result, error) {
	var f struct {
		GoLog   int              `json:"golog"`
		Kind    string           `json:"kind"`
		Name    string           `json:"name"`
		Content model.LogContent `json:"content"`
	}
	if err := json.Unmarshal(data, &f); err != nil || f.GoLog == 0 || f.Kind != "log" {
		return nil, fmt.Errorf("this is not a GoLog file")
	}
	if f.GoLog > 1 {
		return nil, fmt.Errorf("this file was saved by a newer GoLog; update GoLog to open it")
	}
	return &sedlog.Result{Format: "golog", Name: f.Name, Content: f.Content, Warnings: []string{}}, nil
}

// statistics analyses the log in the body (the one on screen, saved or not).
func (s *Server) statistics(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Content  model.LogContent `json:"content"`
		Shuffles int              `json:"shuffles"`
	}
	if !decode(w, r, &b) {
		return
	}
	if err := b.Content.Validate(); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	n := b.Shuffles
	if n <= 0 {
		n = stats.DefaultShuffles
	}
	n = min(n, 20000)
	// Each shuffle costs about as much as the number of beds; keep a long log to a few seconds
	// (10,000 beds and 5000 shuffles take about 2 s).
	const budget = 50_000_000
	cut := false
	if beds := len(b.Content.Beds); beds > 0 && n*beds > budget {
		n, cut = max(500, budget/beds), true
	}
	res := stats.Analyse(b.Content, n)
	if cut {
		res.Notes = append(res.Notes, fmt.Sprintf("This log is long, so it was compared with %d shuffled sections instead of %d.", n, stats.DefaultShuffles))
	}
	reply(w, http.StatusOK, res)
}

func (s *Server) getLog(w http.ResponseWriter, r *http.Request) {
	l, err := s.store.Log(r.Context(), r.PathValue("id"))
	if s.storeErr(w, err, "log") {
		return
	}
	reply(w, http.StatusOK, l)
}

// updateLog saves a whole log. The body must carry the revision it was based on;
// if someone else saved in between, the reply is 409 with the current log.
func (s *Server) updateLog(w http.ResponseWriter, r *http.Request) {
	var l model.Log
	if !decode(w, r, &l) {
		return
	}
	l.ID = r.PathValue("id")
	name, err := model.CleanName(l.Name)
	if err != nil {
		fail(w, http.StatusBadRequest, "log "+err.Error())
		return
	}
	l.Name = name
	if err := l.LogContent.Validate(); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	err = s.store.UpdateLog(r.Context(), &l)
	if errors.Is(err, store.ErrConflict) {
		current, err := s.store.Log(r.Context(), l.ID)
		if err != nil {
			internal(w, err)
			return
		}
		reply(w, http.StatusConflict, map[string]any{"error": store.ErrConflict.Error(), "current": current})
		return
	}
	if s.storeErr(w, err, "log") {
		return
	}
	saved, err := s.store.Log(r.Context(), l.ID)
	if err != nil {
		internal(w, err)
		return
	}
	reply(w, http.StatusOK, saved)
}

func (s *Server) deleteLog(w http.ResponseWriter, r *http.Request) {
	if s.storeErr(w, s.store.DeleteLog(r.Context(), r.PathValue("id")), "log") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// static serves the web app. Unknown paths get index.html so links into the app work.
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		fail(w, http.StatusNotFound, "no such API call")
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if p == "" {
		p = "index.html"
	}
	if st, err := fs.Stat(s.web, p); err != nil || st.IsDir() {
		p = "index.html"
	}
	if strings.HasPrefix(p, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // file names contain a content hash
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeFileFS(w, r, s.web, p)
}

// storeErr writes a reply for a store error and reports whether there was one.
func (s *Server) storeErr(w http.ResponseWriter, err error, what string) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, store.ErrNotFound):
		fail(w, http.StatusNotFound, what+" not found")
	default:
		internal(w, err)
	}
	return true
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		fail(w, http.StatusBadRequest, "could not read the request: "+err.Error())
		return false
	}
	return true
}

func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, msg string) {
	reply(w, status, map[string]string{"error": msg})
}

func internal(w http.ResponseWriter, err error) {
	log.Printf("error: %v", err)
	fail(w, http.StatusInternalServerError, "something went wrong in GoLog; see the terminal for details")
}
