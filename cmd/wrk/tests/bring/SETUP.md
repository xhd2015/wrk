# Scenario

**Feature**: wrk --bring always materializes external dep worktree; Go replace is best-effort

```
# sole external-dep worktree mode: external/{basename}-{token}-{date} under parent; branch {token}-{date}[-N] on dep repo
# parent = git toplevel when cwd is git; parent = abs(cwd) when cwd is plain non-git
# soft SKIP (exit 0) when not a go module / no modules / not a dep / non-git cwd
# git consumer: worktree + /external gitignore + abs path when worktree succeeds
# non-git cwd: worktree under {cwd}/external/; soft-skip replace; skip ensureGitignoreExternal
# --no-dep: worktree (+ gitignore when git parent) only; skip module analyse (no SKIP)
# -v: log major git + go mod tidy pre-line; stream worktree add + tidy child to stderr
consumer (git or plain non-git) + dep path -> wrk --bring <dep> [--no-dep] [-v]
  -> external worktree under {parent}/external/
  -> optional replace+tidy on module match (skipped with --no-dep or non-git)
  -> SKIP local dep replacement: <reason> on stderr when soft-fail (not with --no-dep)
  -> stdout: <external-abs>\n
```

## Preconditions

- Git and Go must be available.
- Dep path must resolve to a usable git repo for worktree create (hard error otherwise).
- **`--bring` does not require consumer git**: non-git cwd still materializes under `{abs(cwd)}/external/` and soft-skips replace + gitignore.
- When cwd **is** git, parent remains the consumer git toplevel (main or linked).
- Module analyse/replace is best-effort only for `--bring` (soft SKIP paths covered under this tree).
- `--no-dep` is long-only and valid only with `--bring`.

## Steps

- Leaves build isolated consumer (git or plain) + dep repos under `req.WorkRoot`.
- `req.RepoDir` is the consumer cwd for `wrk --bring`; `req.ConsumerTop` is the external parent (git toplevel or abs plain cwd).
- `req.Args = []string{"--bring", depPath}` (plus optional `--exec …`, `--no-dep`, `-v`).
- Shared helpers: `initBringConsumerRepo`, `initBringDepRepo`, `bringExternalWorktreePath`, …; see `not-git-cwd/` for plain-cwd fixtures. Multi-module / basename leaves live under this tree.

## Context

- SKIP notices (stderr substrings):
  - `SKIP local dep replacement: <depPath> is not a go module`
  - `SKIP local dep replacement: consumer has no Go modules`
  - `SKIP local dep replacement: <depPath> is not a dependency of any consumer module`
  - `SKIP local dep replacement: <abs-cwd> is not a git repository` (non-git consumer cwd; soft for `--bring` only)
- Mutually exclusive with other wrk modes (same exclusive set as other mode flags).
- `--exec` after successful `--bring` (including SKIP) runs in the external worktree.
- See `not-git-cwd/`, `no-dep/`, `verbose/`, `help-mentions-no-dep/`, and `removed-flags/` for flag/cwd-specific leaves.
- Hard removal of legacy `--dep` / `--all-deps` is asserted under `removed-flags/` (expect RED until implementer lands).
- **Multi-bring** (`Varargs` `--bring p1 p2` and repeat `--bring p1 --bring p2`, fail-fast hard errors, soft-SKIP continue, reject exclusive-multi+`--exec` / exact duplicates; bare `--bring` requires a value): leaves under `multi/`. Compose with create lives under `create-bring/`.

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

	"github.com/xhd2015/doctest/assert"
)

const bringDepModulePath = "example.com/dep"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	ensureBringHelpersUsed()
	return nil
}

type bringGoModJSON struct {
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

func readBringGoMod(modDir string) (*bringGoModJSON, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var mod bringGoModJSON
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func bringGitignoreContainsExternal(top string) (bool, error) {
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

func initBringConsumerRepo(t *testing.T, workRoot string, withRequire bool) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	if withRequire {
		runBringGo(t, consumer, "mod", "edit", "-require="+bringDepModulePath+"@v0.0.0")
	}
	// Canonicalize so path comparisons match git's toplevel canonicalization
	// (e.g. macOS /var -> /private/var). No-op on filesystems without symlinks.
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	return consumer
}

func initBringDepRepo(t *testing.T, workRoot, name string, withGoMod bool) string {
	t.Helper()
	dep := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, dep)
	if withGoMod {
		writeFile(t, filepath.Join(dep, "go.mod"), "module "+bringDepModulePath+"\n\ngo 1.22\n")
		writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n")
		runGitIsolated(t, dep, "add", "go.mod", "dep.go")
		runGitIsolated(t, dep, "commit", "-m", "add go module")
	}
	// Canonicalize for SKIP messages that include resolved dep path.
	dep, err := filepath.EvalSymlinks(dep)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", dep, err)
	}
	return dep
}

func runBringGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func bringExternalWorktreePath(consumerTop, depBasename, token string, suffix int) string {
	name := fmt.Sprintf("%s-%s-%s", depBasename, token, wrkDate)
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return filepath.Join(consumerTop, "external", name)
}

func bringHasReplaceForModule(mod *bringGoModJSON, modulePath, wantPath string) bool {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath && repl.New.Path == wantPath {
			return true
		}
	}
	return false
}

func bringHasAnyReplaceForModule(mod *bringGoModJSON, modulePath string) bool {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath {
			return true
		}
	}
	return false
}

// assertBringPathThenChildStdout expects mode path line then child stdout line (e.g. pwd).
func assertBringPathThenChildStdout(t *testing.T, stdout, wantPath, childLine string) {
	t.Helper()
	body := wantPath + "\n" + childLine
	assert.Output(t, stdout, v2StdoutTemplate(body))
}


func bringCountGitignoreExternalLines(top string) (int, error) {
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

// bringReadWorktreeMainRepo reads a linked worktree's .git file and returns the main
// repository path it is registered under (i.e. the <mainRepo> in
// "<mainRepo>/.git/worktrees/<name>"). The result is symlink-canonicalized so
// comparisons match git's resolved paths (macOS /var -> /private/var).
func bringReadWorktreeMainRepo(wtDir string) (string, error) {
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

func ensureBringHelpersUsed() {
	_ = readBringGoMod
	_ = bringGitignoreContainsExternal
	_ = bringCountGitignoreExternalLines
	_ = bringReadWorktreeMainRepo
	_ = initBringConsumerRepo
	_ = initBringDepRepo
	_ = runBringGo
	_ = bringExternalWorktreePath
	_ = bringHasReplaceForModule
	_ = bringHasAnyReplaceForModule
	_ = assertBringPathThenChildStdout
}
```
