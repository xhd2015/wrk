# Scenario

**Feature**: `--bring` after a create form applies inside the new (or Policy B reused) worktree

```
# create compose: parse → preflightResolveBringArgs → window → create → UX ∥ bring → exec → cd
wrk src --bring d1 d2          -> new WT under WRK_HOME; bring into that WT
wrk --new --bring d1           -> explicit create from cwd; bring into new WT
wrk src dst --bring d1         -> spawn at dst; bring into dst
# preflight fail → no new WT; apply fail after create → keep WT, no rollback
src (git Go module) + dep git modules -> wrk <create-form> --bring <deps>
  -> create path on stdout; external abs paths under the **new** WT
  -> source src/go.mod and src/external untouched
```

## Preconditions

- Git + Go on PATH.
- L2: every leaf sets `req.InProcess = true`.
- Do **not** redefine root `Request` / `Response` / `Run`.
- Isolated `{WRK_HOME}` / `WRK_DATE=2026-06-30` via harness.
- Compose leaves pass `--no-config` unless they test UX flags.

## Steps

- Shared fixtures: source repo `src` (git Go module requiring dep module(s));
  dep repos `mydep1` / `mydep2` under `req.WorkRoot`.
- Field mapping:
  - `req.TargetDir` — first positional `src` when running from `WorkRoot`
  - `req.SpawnDir` — `<target-dir>` `dst` when set
  - `req.RepoDir` — process cwd (`WorkRoot` for `wrk src …`; `src` for `wrk --new`)
  - `req.DepPath` / `req.SecondRepo` — dep main paths
  - `req.ConsumerTop` — expected bring consumer (new WT / spawn path)
  - `req.TaskDesc` / `req.TaskFlag` — `-t` / `--task`
  - extra flags in `req.Args` (`--no-config`, `--bring …`, `--new`, `--exec`, `--no-dep`)

## Context

- Exclusive `--bring` (no create form) still brings into **cwd** (`bring/basic`).
- `wrk --bring x1 x2 src` treats `src` as a dep path, not a project.
- `--exec` after create+bring: `cmd.Dir` is the **project** worktree, not external.
- Event `command` is `"create"` (not `"bring"`); `args` include `--bring` and every path.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	createBringSrcModule  = "example.com/src"
	createBringDep1Module = "example.com/dep1"
	createBringDep2Module = "example.com/dep2"
	createBringSrcName    = "src"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	ensureCreateBringHelpersUsed()
	return nil
}

type createBringGoModJSON struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
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

type createBringEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func runCreateBringGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func initCreateBringSrc(t *testing.T, workRoot string, requireModules ...string) string {
	t.Helper()
	src := filepath.Join(workRoot, createBringSrcName)
	initGitRepoOnMain(t, src)
	writeFile(t, filepath.Join(src, "go.mod"), "module "+createBringSrcModule+"\n\ngo 1.22\n")
	for _, mod := range requireModules {
		runCreateBringGo(t, src, "mod", "edit", "-require="+mod+"@v0.0.0")
	}
	src, err := filepath.EvalSymlinks(src)
	if err != nil {
		t.Fatalf("eval symlinks src: %v", err)
	}
	return src
}

func initCreateBringDep(t *testing.T, workRoot, name, modulePath string) string {
	t.Helper()
	dep := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, dep)
	writeFile(t, filepath.Join(dep, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n")
	runGitIsolated(t, dep, "add", "go.mod", "dep.go")
	runGitIsolated(t, dep, "commit", "-m", "add "+name+" module")
	dep, err := filepath.EvalSymlinks(dep)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", name, err)
	}
	return dep
}

func createBringExternalPath(consumerTop, depBasename string) string {
	return filepath.Join(consumerTop, "external", depBasename)
}

func createBringDefaultWT(req *Request) string {
	return worktreePath(req.WrkHome, createBringSrcName, "main", wrkDate, 0)
}

func createBringDefaultWTWithTask(req *Request, task string) string {
	return worktreePathWithTask(req.WrkHome, createBringSrcName, "main", wrkDate, slugify(task), 0)
}

func readCreateBringGoMod(modDir string) (*createBringGoModJSON, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var mod createBringGoModJSON
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func createBringHasReplace(mod *createBringGoModJSON, modulePath, wantPath string) bool {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath && repl.New.Path == wantPath {
			return true
		}
	}
	return false
}

func createBringHasAnyReplace(mod *createBringGoModJSON, modulePath string) bool {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath {
			return true
		}
	}
	return false
}

func createBringGitignoreHasExternal(top string) (bool, error) {
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

func createBringSnapshotGoMod(t *testing.T, req *Request, modDir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		t.Fatalf("snapshot go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(req.WorkRoot, "go.mod.before"), data, 0o644); err != nil {
		t.Fatalf("write go.mod.before: %v", err)
	}
}

func createBringAssertGoModUnchanged(t *testing.T, req *Request, modDir string) {
	t.Helper()
	before, err := os.ReadFile(filepath.Join(req.WorkRoot, "go.mod.before"))
	if err != nil {
		t.Fatalf("read go.mod.before: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("source go.mod changed\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func createBringListHomeWTs(t *testing.T, req *Request) []string {
	t.Helper()
	root := filepath.Join(req.WrkHome, "worktrees")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("readdir worktrees: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func createBringStdoutHasLine(stdout, want string) bool {
	for _, line := range strings.Split(stdout, "\n") {
		if line == want {
			return true
		}
	}
	return false
}

func createBringLastEvent(t *testing.T, wrkHome string) createBringEvent {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(wrkHome, "events.jsonl"))
	if err != nil {
		t.Fatalf("read events.jsonl: %v", err)
	}
	var last createBringEvent
	n := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev createBringEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		last = ev
		n++
	}
	if n == 0 {
		t.Fatal("expected at least one events.jsonl entry")
	}
	return last
}

func createBringArgsContain(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

func ensureCreateBringHelpersUsed() {
	_ = initCreateBringSrc
	_ = initCreateBringDep
	_ = createBringExternalPath
	_ = createBringDefaultWT
	_ = createBringDefaultWTWithTask
	_ = readCreateBringGoMod
	_ = createBringHasReplace
	_ = createBringHasAnyReplace
	_ = createBringGitignoreHasExternal
	_ = createBringSnapshotGoMod
	_ = createBringAssertGoModUnchanged
	_ = createBringListHomeWTs
	_ = createBringStdoutHasLine
	_ = createBringLastEvent
	_ = createBringArgsContain
	_ = createBringSrcModule
	_ = createBringDep1Module
	_ = createBringDep2Module
}
```
