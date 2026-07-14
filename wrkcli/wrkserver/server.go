package wrkserver

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Options configures a wrkserver instance.
type Options struct {
	// WrkHome is the wrk storage root (projects.json, worktrees/).
	// Empty resolves like the CLI (WRK_HOME or ~/.wrk).
	WrkHome string
}

// Server holds HTTP handlers for wrk project list and worktree create.
type Server struct {
	wrkHome string // may be empty until resolved per-request / at New
}

// New constructs a Server from Options.
func New(opts Options) *Server {
	return &Server{wrkHome: strings.TrimSpace(opts.WrkHome)}
}

// Register mounts fixed leaves under base (trailing slash optional).
//
//	GET  {base}/projects           → ListProjects
//	POST {base}/worktrees          → CreateWorktree
//	POST {base}/ops                → CreateOp (mock streaming ops)
//	GET  {base}/ops/{id}/logs      → StreamOpLogs (SSE)
//
// base is host-owned; wrkserver does not hardcode /api/wrk.
func (s *Server) Register(mux *http.ServeMux, base string) {
	base = strings.TrimRight(base, "/")
	mux.HandleFunc(base+"/projects", s.ListProjects)
	mux.HandleFunc(base+"/worktrees", s.CreateWorktree)
	mux.HandleFunc(base+"/ops", s.CreateOp)
	// Go 1.22+ path value for SSE log stream.
	mux.HandleFunc(base+"/ops/{id}/logs", s.StreamOpLogs)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorBody{Error: msg})
}
