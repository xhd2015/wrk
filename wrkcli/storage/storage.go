package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

const (
	SourceAuto   = "auto"
	SourceManual = "manual"
)

// Project is one recorded main repository entry.
type Project struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

// ProjectsFile is the on-disk projects.json schema.
type ProjectsFile struct {
	Version  int       `json:"version"`
	Projects []Project `json:"projects"`
}

// Event is one append-only events.jsonl record.
type Event struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

// NormalizePath returns an absolute path, resolving symlinks when possible.
func NormalizePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

func projectsPath(wrkHome string) string {
	return filepath.Join(wrkHome, "projects.json")
}

func eventsPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

// LoadProjects reads projects.json, returning an empty file when absent.
func LoadProjects(wrkHome string) (ProjectsFile, error) {
	data, err := os.ReadFile(projectsPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return ProjectsFile{}, nil
		}
		return ProjectsFile{}, err
	}
	var pf ProjectsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return ProjectsFile{}, fmt.Errorf("parse projects.json: %w", err)
	}
	return pf, nil
}

// SaveProjects writes projects.json under wrkHome.
func SaveProjects(wrkHome string, pf ProjectsFile) error {
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		return fmt.Errorf("create wrk home: %w", err)
	}
	data, err := json.Marshal(pf)
	if err != nil {
		return err
	}
	return os.WriteFile(projectsPath(wrkHome), data, 0o644)
}

func findProject(pf ProjectsFile, path string) bool {
	norm := NormalizePath(path)
	for _, p := range pf.Projects {
		if NormalizePath(p.Path) == norm {
			return true
		}
	}
	return false
}

// RecordProject appends a project when absent. Re-adding is idempotent; the
// first recorded source wins.
func RecordProject(wrkHome, path, source string) error {
	path = NormalizePath(path)
	pf, err := LoadProjects(wrkHome)
	if err != nil {
		return err
	}
	if findProject(pf, path) {
		return nil
	}
	pf.Version = 1
	pf.Projects = append(pf.Projects, Project{
		Path:    path,
		AddedAt: time.Now().UTC().Format(time.RFC3339),
		Source:  source,
	})
	return SaveProjects(wrkHome, pf)
}

// RemoveProject deletes the project entry matching normalized path. Returns
// true when an entry was removed, false when no matching entry exists.
func RemoveProject(wrkHome, path string) (bool, error) {
	path = NormalizePath(path)
	pf, err := LoadProjects(wrkHome)
	if err != nil {
		return false, err
	}
	var kept []Project
	removed := false
	for _, p := range pf.Projects {
		if NormalizePath(p.Path) == path {
			removed = true
			continue
		}
		kept = append(kept, p)
	}
	if !removed {
		return false, nil
	}
	pf.Projects = kept
	return true, SaveProjects(wrkHome, pf)
}

// ListProjects returns recorded project paths sorted lexicographically.
func ListProjects(wrkHome string) ([]string, error) {
	pf, err := LoadProjects(wrkHome)
	if err != nil {
		return nil, err
	}
	paths := make([]string, len(pf.Projects))
	for i, p := range pf.Projects {
		paths[i] = NormalizePath(p.Path)
	}
	sort.Strings(paths)
	return paths, nil
}

// ResetEventsIfDoctest truncates events.jsonl when DOCTEST_SESSION_ID is set so
// doctest assertions observe only the current invocation's event.
func ResetEventsIfDoctest(wrkHome string) error {
	if os.Getenv("DOCTEST_SESSION_ID") == "" {
		return nil
	}
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		return fmt.Errorf("create wrk home: %w", err)
	}
	f, err := os.OpenFile(eventsPath(wrkHome), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

// AppendEvent appends one JSON line to events.jsonl.
func AppendEvent(wrkHome string, ev Event) error {
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		return fmt.Errorf("create wrk home: %w", err)
	}
	if ev.Args == nil {
		ev.Args = []string{}
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(eventsPath(wrkHome), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

// ResolveMainRepoForWorkDir returns the main repo for workDir when it exists and
// is inside a git work tree.
func ResolveMainRepoForWorkDir(workDir string) (string, bool) {
	if _, err := os.Stat(workDir); err != nil {
		return "", false
	}
	if !worktree.IsInsideWorkTree(workDir) {
		return "", false
	}
	top, err := worktree.ShowToplevel(workDir)
	if err != nil {
		return "", false
	}
	main, err := worktree.ResolveMainRepo(top)
	if err != nil {
		return "", false
	}
	return NormalizePath(main), true
}

// AutoRecord records the main repo for workDir with source "auto" when git
// resolves successfully.
func AutoRecord(wrkHome, workDir string) (string, error) {
	main, ok := ResolveMainRepoForWorkDir(workDir)
	if !ok {
		return "", nil
	}
	if err := RecordProject(wrkHome, main, SourceAuto); err != nil {
		return main, err
	}
	return main, nil
}