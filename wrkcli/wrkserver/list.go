package wrkserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/wrk/wrkcli"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// ListProjects handles GET {base}/projects.
func (s *Server) ListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	wrkHome, err := wrkcli.ResolveWrkHome(s.wrkHome)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Always emit an array (never null) for the projects key.
	projects := make([]ProjectStatus, 0, len(paths))
	for _, p := range paths {
		projects = append(projects, gatherOneProject(p))
	}
	writeJSON(w, http.StatusOK, ListProjectsResponse{Projects: projects})
}

func gatherOneProject(mainPath string) ProjectStatus {
	mainPath = storage.NormalizePath(mainPath)
	ps := ProjectStatus{
		Path:      mainPath,
		Name:      filepath.Base(mainPath),
		Worktrees: []WorktreeStatus{},
	}

	if _, err := os.Stat(mainPath); err != nil {
		if os.IsNotExist(err) {
			ps.Error = "path does not exist"
		} else {
			ps.Error = err.Error()
		}
		return ps
	}

	if !worktree.IsInsideWorkTree(mainPath) {
		ps.Error = fmt.Sprintf("%s is not a git repository", mainPath)
		return ps
	}

	// Prefer resolved main repo path for status (registry may point at a checkout).
	if top, err := worktree.ShowToplevel(mainPath); err == nil {
		if main, err := worktree.ResolveMainRepo(top); err == nil {
			mainPath = storage.NormalizePath(main)
			ps.Path = mainPath
			ps.Name = filepath.Base(mainPath)
		}
	}

	branch, err := worktree.ReadBranch(mainPath)
	if err != nil {
		ps.Error = err.Error()
		return ps
	}
	ps.Branch = branch

	short, subject, err := commitShortSubject(mainPath)
	if err != nil {
		ps.Error = err.Error()
		return ps
	}
	ps.Commit = short
	ps.Subject = subject

	linked, err := worktree.ListLinked(mainPath)
	if err != nil {
		ps.Error = err.Error()
		return ps
	}

	// Main clean: skip untracked paths that are nested linked worktrees.
	clean, err := mainIsClean(mainPath, linked)
	if err != nil {
		ps.Error = err.Error()
		return ps
	}
	ps.Clean = clean

	for _, entry := range linked {
		wt := WorktreeStatus{
			Path:   storage.NormalizePath(entry.Path),
			Name:   filepath.Base(entry.Path),
			Branch: entry.Branch,
			IsMain: false,
		}
		if worktree.IsDead(entry.Path) {
			wt.Error = "path does not exist"
			ps.Worktrees = append(ps.Worktrees, wt)
			continue
		}
		isClean, err := worktree.IsCleanWrk(entry.Path)
		if err != nil {
			wt.Error = err.Error()
		} else {
			wt.Clean = isClean
		}
		ps.Worktrees = append(ps.Worktrees, wt)
	}
	return ps
}

func commitShortSubject(repoPath string) (short, subject string, err error) {
	out, err := cmd.Run(context.Background(), repoPath, "log", "-1", "--pretty=format:%h %s")
	if err != nil {
		return "", "", err
	}
	out = strings.TrimSpace(out)
	if i := strings.IndexByte(out, ' '); i >= 0 {
		return out[:i], out[i+1:], nil
	}
	return out, "", nil
}

func mainIsClean(mainPath string, linked []worktree.Entry) (bool, error) {
	out, err := cmd.Run(context.Background(), mainPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	skip := skipUntrackedRelPaths(mainPath, linked)
	if len(skip) == 0 {
		return worktree.IsCleanWrk(mainPath)
	}
	var filtered strings.Builder
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			path := strings.TrimSpace(line[2:])
			path = strings.TrimSuffix(path, "/")
			if _, ok := skip[path]; ok {
				continue
			}
		}
		if filtered.Len() > 0 {
			filtered.WriteByte('\n')
		}
		filtered.WriteString(line)
	}
	return strings.TrimSpace(filtered.String()) == "", nil
}

func skipUntrackedRelPaths(mainRepo string, linked []worktree.Entry) map[string]struct{} {
	skip := make(map[string]struct{})
	for _, entry := range linked {
		if worktree.IsDead(entry.Path) {
			continue
		}
		rel, err := filepath.Rel(mainRepo, entry.Path)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "..") {
			continue
		}
		skip[rel] = struct{}{}
	}
	return skip
}
