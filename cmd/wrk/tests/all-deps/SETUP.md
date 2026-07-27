# Scenario

**Feature**: wrk --all-deps discovers deps from registered projects in projects.json

```
# consumer requires deps; registered projects in WRK_HOME/projects.json -> wrk --all-deps links each match
consumer (go.mod + git) + projects.json (dep main repos) -> wrk --all-deps -> stdout one line per dep + summary
```

## Preconditions

- Git and Go must be available.
- Consumer cwd must be inside a git work tree with a `go.mod` (or sub-module go.mod).
- Dep repos registered in `projects.json` must be git main repos on branch `main` with committed module trees.

## Steps

- Tests build an isolated consumer git repo plus dep repos at arbitrary paths under `{WorkRoot}`.
- Dep repos are registered via `wrk --add` or pre-seeded `projects.json` (never via `--scan-root`).
- `req.RepoDir` is the consumer cwd for `wrk --all-deps`.
- `req.Args = []string{"--all-deps"}` (or with `--dry-run` for planning leaves).

## Context

- Output order follows **lexicographic project path order** (same as `wrk --projects`).
- Empty or absent `projects.json` → `wrked 0 deps`, exit 0, no `external/`, no replaces.
- `--scan-root` and `WRK_SCAN_ROOT` are removed; passing `--scan-root` must error.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

// allDepsGoModJSON mirrors the go.mod structure read via `go mod edit -json`.
type allDepsGoModJSON struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"New"`
	} `json:"Replace"`
}

type allDepsProjectEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type allDepsProjectsFile struct {
	Version  int                   `json:"version"`
	Projects []allDepsProjectEntry `json:"projects"`
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	return nil
}

func allDepsProjectsJSONPath(wrkHome string) string {
	return filepath.Join(wrkHome, "projects.json")
}

func allDepsResolvePath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}

// registerAllDepsProject records a main repo in projects.json via wrk --add.
func registerAllDepsProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, req.WorkRoot, "--add", repoPath)
}

// registerAllDepsProjects registers multiple repos in lexicographic path order
// (registration order does not affect runtime order; ListProjects sorts).
func registerAllDepsProjects(t *testing.T, req *Request, repoPaths ...string) {
	t.Helper()
	sorted := append([]string(nil), repoPaths...)
	sort.Slice(sorted, func(i, j int) bool {
		return allDepsResolvePath(t, sorted[i]) < allDepsResolvePath(t, sorted[j])
	})
	for _, p := range sorted {
		registerAllDepsProject(t, req, p)
	}
}

// writeAllDepsProjectsJSON seeds projects.json with the given paths (for missing-path
// or non-git scenarios where wrk --add would fail or is inappropriate).
func writeAllDepsProjectsJSON(t *testing.T, wrkHome string, paths ...string) {
	t.Helper()
	var projects []allDepsProjectEntry
	for _, p := range paths {
		projects = append(projects, allDepsProjectEntry{
			Path:    p,
			AddedAt: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Source:  "manual",
		})
	}
	pf := allDepsProjectsFile{Version: 1, Projects: projects}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		t.Fatalf("marshal projects.json: %v", err)
	}
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		t.Fatalf("mkdir WRK_HOME: %v", err)
	}
	if err := os.WriteFile(allDepsProjectsJSONPath(wrkHome), append(data, '\n'), 0644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

// allDepsDepDir returns the conventional dep repo path under workRoot/deps/<name>.
func allDepsDepDir(workRoot, name string) string {
	return filepath.Join(workRoot, "deps", name)
}

// allDepsReadGoMod reads a directory's go.mod via `go mod edit -json`.
func allDepsReadGoMod(modDir string) (*allDepsGoModJSON, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var mod allDepsGoModJSON
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

// allDepsHasReplaceForModule reports whether go.mod has a replace for
// modulePath whose new path matches wantPath (empty wantPath → any path).
func allDepsHasReplaceForModule(mod *allDepsGoModJSON, modulePath, wantPath string) bool {
	for _, repl := range mod.Replace {
		if repl.Old.Path != modulePath {
			continue
		}
		if wantPath == "" || repl.New.Path == wantPath {
			return true
		}
	}
	return false
}

// allDepsReplacePathForModule returns the current replace new-path for
// modulePath, or "" if none.
func allDepsReplacePathForModule(mod *allDepsGoModJSON, modulePath string) string {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath {
			return repl.New.Path
		}
	}
	return ""
}

// allDepsGitignoreContainsExternal reports whether .gitignore has a `/external` line.
func allDepsGitignoreContainsExternal(top string) (bool, error) {
	data, err := os.ReadFile(filepath.Join(top, ".gitignore"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "/external" {
			return true, nil
		}
	}
	return false, nil
}

// allDepsCountGitignoreExternalLines counts `/external` lines in .gitignore.
func allDepsCountGitignoreExternalLines(top string) (int, error) {
	data, err := os.ReadFile(filepath.Join(top, ".gitignore"))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "/external" {
			n++
		}
	}
	return n, nil
}

// allDepsExternalRelPath returns the relative `./external/<name>` form printed
// by wrk for a given dep basename (token "main", date wrkDate, no suffix).
func allDepsExternalRelPath(depBasename string) string {
	return fmt.Sprintf("./external/%s-main-%s", depBasename, wrkDate)
}

// allDepsExternalAbsPath returns the absolute external worktree path for a dep basename.
func allDepsExternalAbsPath(consumerTop, depBasename string) string {
	resolved, err := filepath.EvalSymlinks(consumerTop)
	if err != nil {
		resolved = consumerTop
	}
	return filepath.Join(resolved, "external", fmt.Sprintf("%s-main-%s", depBasename, wrkDate))
}

// allDepsDepMainRepo returns the resolved main-repo path of a dep repo.
func allDepsDepMainRepo(depPath string) string {
	resolved, err := filepath.EvalSymlinks(depPath)
	if err != nil {
		return depPath
	}
	return resolved
}

// allDepsRunGo runs a go command in dir, failing the test on error.
func allDepsRunGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// initAllDepsRepo creates a git repo on branch main at path with a committed
// go.mod (module modulePath) and one .go file (package pkgName).
func initAllDepsRepo(t *testing.T, path, modulePath, pkgName string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, pkgName+".go"), "package "+pkgName+"\n")
	runGitIsolated(t, path, "add", "go.mod", pkgName+".go")
	runGitIsolated(t, path, "commit", "-m", "init "+modulePath)
}

// initAllDepsConsumer creates a consumer git repo on main with a go.mod that
// requires the given module paths. extraGoMod (if non-empty) is appended to
// go.mod verbatim (used for pre-existing replace directives).
func initAllDepsConsumer(t *testing.T, workRoot string, requires []string, extraGoMod string) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	mkdirAll(t, consumer)
	runGitIsolated(t, consumer, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, consumer, "config", "user.email", "test@test.com")
	runGitIsolated(t, consumer, "config", "user.name", "Test")
	var sb strings.Builder
	sb.WriteString("module example.com/consumer\n\ngo 1.22\n")
	for _, m := range requires {
		sb.WriteString("\nrequire " + m + " v0.0.0\n")
	}
	if extraGoMod != "" {
		sb.WriteString("\n" + extraGoMod + "\n")
	}
	writeFile(t, filepath.Join(consumer, "go.mod"), sb.String())
	writeFile(t, filepath.Join(consumer, "main.go"), "package main\n")
	runGitIsolated(t, consumer, "add", "go.mod", "main.go")
	runGitIsolated(t, consumer, "commit", "-m", "init consumer")
	resolved, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		return consumer
	}
	return resolved
}

// nestedExternalAbsSubPath returns the resolved absolute path to a sub-module
// directory inside the external worktree of repoBasename.
func nestedExternalAbsSubPath(consumerTop, repoBasename, subdir string) string {
	resolved, err := filepath.EvalSymlinks(consumerTop)
	if err != nil {
		resolved = consumerTop
	}
	return filepath.Join(resolved, "external", fmt.Sprintf("%s-main-%s", repoBasename, wrkDate), subdir)
}

// nestedExternalRelSubPath returns the relative ./external/.../<subdir> form
// printed by wrk for a nested sub-module.
func nestedExternalRelSubPath(repoBasename, subdir string) string {
	return fmt.Sprintf("./external/%s-main-%s/%s", repoBasename, wrkDate, subdir)
}

func allDepsStdoutV2(body string) string {
	return v2StdoutTemplate(body)
}

func allDepsEnsureHelpersUsed() {
	_ = allDepsGoModJSON{}
	_ = allDepsProjectEntry{}
	_ = allDepsProjectsFile{}
	_ = allDepsProjectsJSONPath
	_ = allDepsResolvePath
	_ = registerAllDepsProject
	_ = registerAllDepsProjects
	_ = writeAllDepsProjectsJSON
	_ = allDepsDepDir
	_ = allDepsReadGoMod
	_ = allDepsHasReplaceForModule
	_ = allDepsReplacePathForModule
	_ = allDepsGitignoreContainsExternal
	_ = allDepsCountGitignoreExternalLines
	_ = allDepsExternalRelPath
	_ = allDepsExternalAbsPath
	_ = allDepsDepMainRepo
	_ = allDepsRunGo

	_ = initAllDepsRepo
	_ = initAllDepsConsumer
	_ = nestedExternalAbsSubPath
	_ = nestedExternalRelSubPath
	_ = allDepsStdoutV2
}
```