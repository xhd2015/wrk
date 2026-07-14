package wrkserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/xhd2015/wrk/wrkcli"
)

// CreateWorktree handles POST {base}/worktrees.
func (s *Server) CreateWorktree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateWorktreeRequest
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}
		if len(strings.TrimSpace(string(body))) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
				return
			}
		}
	}

	projectPath := strings.TrimSpace(req.ProjectPath)
	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "project_path is required")
		return
	}

	// Whitespace-only / omitted task → no slug (UI-friendly; differs from CLI --task "").
	slug, err := wrkcli.NormalizeTaskSlug(req.Task)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	wrkHome, err := wrkcli.ResolveWrkHome(s.wrkHome)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result, err := wrkcli.CreateDefaultWorktree(projectPath, wrkHome, slug)
	if err != nil {
		// Create failures from missing git / bad paths are client errors.
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CreateWorktreeResponse{
		Path:   result.Path,
		Branch: result.Branch,
	})
}
