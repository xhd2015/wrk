# Scenario

**Feature**: wrk --bring always materializes external dep worktree; Go replace is best-effort

```
# like --dep: external/{basename}-{token}-{date} under consumer; branch {token}-{date}[-N] on dep repo
# unlike --dep: soft SKIP (exit 0) when not a go module / consumer has no modules / not a dependency
# always: worktree + /external gitignore + abs path on stdout when git/worktree succeeds
# --no-dep: worktree + gitignore only; skip module match scan + replace + go mod tidy
# -v: log major git + go mod tidy pre-line; stream worktree add + tidy child to stderr
consumer (git) + dep path -> wrk --bring <dep> [--no-dep] [-v]
  -> external worktree under consumer/external/
  -> optional replace+tidy on module match (skipped with --no-dep)
  -> SKIP local dep replacement: <reason> on stderr when soft-fail (not with --no-dep)
  -> stdout: <external-abs>\n
```

## Preconditions

- Git and Go must be available.
- Consumer cwd must be inside a usable git work tree (main or linked).
- Dep path must resolve to a usable git repo for worktree create (hard error otherwise).
- Module analyse/replace is best-effort only for `--bring` (strict hard errors remain for `--dep`).
- `--no-dep` is long-only and valid only with `--bring` / `--dep` / `--all-deps`.

## Steps

- Leaves build isolated consumer + dep repos under `req.WorkRoot`.
- `req.RepoDir` is the consumer cwd for `wrk --bring`.
- `req.Args = []string{"--bring", depPath}` (plus optional `--exec …`, `--no-dep`, `-v`).
- Shared helpers mirror `dep/` fixture patterns (`initConsumerRepo`, `initDepRepo`, `externalWorktreePath`, …).

## Context

- SKIP notices (stderr substrings):
  - `SKIP local dep replacement: <depPath> is not a go module`
  - `SKIP local dep replacement: consumer has no Go modules`
  - `SKIP local dep replacement: <depPath> is not a dependency of any consumer module`
- Mutually exclusive with `--dep` (and the same exclusive mode set as `--dep`).
- `--exec` after successful `--bring` (including SKIP) runs in the external worktree.
- See `no-dep/`, `verbose/`, and `help-mentions-no-dep/` for flag-specific leaves.

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

func ensureBringHelpersUsed() {
	_ = readBringGoMod
	_ = bringGitignoreContainsExternal
	_ = initBringConsumerRepo
	_ = initBringDepRepo
	_ = runBringGo
	_ = bringExternalWorktreePath
	_ = bringHasReplaceForModule
	_ = bringHasAnyReplaceForModule
	_ = assertBringPathThenChildStdout
}
```
