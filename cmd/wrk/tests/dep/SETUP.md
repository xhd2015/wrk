# Scenario

**Feature**: wrk --dep spawns external dependency worktree and edits consumer go.mod

```
# consumer requires dep; wrk --dep <dep-repo> -> external/<name> worktree + replace + tidy + gitignore
# --no-dep: strict pre-checks still apply; then worktree + gitignore only (no replace/tidy)
consumer (go.mod + git) + dep repo -> wrk --dep [--no-dep] -> stdout external path
# path: external/{depBasename}-{token}-{date}[-N]
# branch: {token}-{date}[-N]  (NO dep basename; always worktree add -b)
```

## Branch naming (P2 behavior-change)

- External **branch** is `{token}-{date}[-N]` — not `{depBasename}-{token}-{date}`.
- Always create a new branch (`-b`); joint path+branch collision against dep main.
- `basic/` asserts no basename on branch; `branch-collision-suffix/` covers preferred-branch pre-exists → `-1`.

## Preconditions

- Git and Go must be available.
- Consumer cwd must be inside a git work tree with a `go.mod`.
- Dep path must be a git repo with a valid Go module listed in consumer go.mod.
- `--no-dep` long-only; valid with `--dep` / `--bring` / `--all-deps` only.

## Steps

- Tests build isolated consumer + dep repos under `req.WorkRoot`.
- `req.RepoDir` is the consumer cwd for `wrk --dep`.
- `req.Args = []string{"--dep", depPath}` (plus optional `--no-dep`).
- See `no-dep/` for worktree-only and strict not-a-dependency leaves.

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

const depModulePath = "example.com/dep"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	return nil
}

type goModJSON struct {
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

func readGoMod(modDir string) (*goModJSON, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var mod goModJSON
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func hasReplaceLocal(modDir string) (bool, error) {
	mod, err := readGoMod(modDir)
	if err != nil {
		return false, err
	}
	for _, repl := range mod.Replace {
		p := repl.New.Path
		if p == "" {
			continue
		}
		if repl.New.Version != "" {
			continue
		}
		if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") || filepath.IsAbs(p) {
			return true, nil
		}
	}
	return false, nil
}

func gitignoreContainsExternal(top string) (bool, error) {
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

func countGitignoreExternalLines(top string) (int, error) {
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

func initConsumerRepo(t *testing.T, workRoot string, withRequire bool) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	if withRequire {
		runGo(t, consumer, "mod", "edit", "-require="+depModulePath+"@v0.0.0")
	}
	// Canonicalize so path comparisons match git's toplevel canonicalization
	// (e.g. macOS /var -> /private/var). No-op on filesystems without symlinks.
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	return consumer
}

func initDepRepo(t *testing.T, workRoot, name string, withGoMod bool) string {
	t.Helper()
	depPath := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, depPath)
	if withGoMod {
		writeFile(t, filepath.Join(depPath, "go.mod"), "module "+depModulePath+"\n\ngo 1.22\n")
		writeFile(t, filepath.Join(depPath, "dep.go"), "package dep\n")
		runGitIsolated(t, depPath, "add", "go.mod", "dep.go")
		runGitIsolated(t, depPath, "commit", "-m", "add go module")
	}
	return depPath
}

func runGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func externalWorktreePath(consumerTop, depBasename, token string, suffix int) string {
	name := fmt.Sprintf("%s-%s-%s", depBasename, token, wrkDate)
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return filepath.Join(consumerTop, "external", name)
}

func hasReplaceForModule(mod *goModJSON, modulePath, wantPath string) bool {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath && repl.New.Path == wantPath {
			return true
		}
	}
	return false
}

// readWorktreeMainRepo reads a linked worktree's .git file and returns the main
// repository path it is registered under (i.e. the <mainRepo> in
// "<mainRepo>/.git/worktrees/<name>"). The result is symlink-canonicalized so
// comparisons match git's resolved paths (macOS /var -> /private/var).
func readWorktreeMainRepo(wtDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(wtDir, ".git"))
	if err != nil {
		return "", fmt.Errorf("read .git file: %w", err)
	}
	s := strings.TrimSpace(string(data))
	const prefix = "gitdir: "
	if !strings.HasPrefix(s, prefix) {
		return "", fmt.Errorf("unexpected .git file format in %s: %s", wtDir, s)
	}
	gitdir := strings.TrimSpace(s[len(prefix):])
	// gitdir is <mainRepo>/.git/worktrees/<name>; climb three dirs to mainRepo.
	mainRepo := filepath.Dir(filepath.Dir(filepath.Dir(gitdir)))
	if resolved, err := filepath.EvalSymlinks(mainRepo); err == nil {
		mainRepo = resolved
	}
	return mainRepo, nil
}
```