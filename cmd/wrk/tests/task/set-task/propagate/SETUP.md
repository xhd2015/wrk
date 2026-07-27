# Scenario

**Feature**: --set-task also updates gitdir metadata for nested linked worktrees found under the consumer

```
# before renaming: discover all repos under consumer via wrk --repos scan
# after rename: update <mainRepo>/.git/worktrees/<name>/gitdir to new path
wrk --set-task "new slug" (inside consumer wt with external dep)
  -> discover nested repos (same scan as wrk --repos)
  -> git worktree move + git branch -m
  -> for each nested worktree: update gitdir in owning main repo
```

## Preconditions

- Git must be available.
- Consumer worktree contains one or more nested linked worktrees (external deps created via `wrk --dep` or any repos discoverable by `wrk --repos`).

## Steps

- Setup creates main consumer repo, spawns consumer worktree, creates dep repos, and runs `wrk --dep` from inside the consumer worktree to materialize external dep worktrees.
- Each leaf variant sets up different nested repo configurations before running `wrk --set-task`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	return nil
}

const depModulePath = "example.com/dep"

func initConsumerRepo(t *testing.T, workRoot string, withRequire bool) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	if withRequire {
		cmd := exec.Command("go", "mod", "edit", "-require="+depModulePath+"@v0.0.0")
		cmd.Dir = consumer
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go mod edit -require: %v\n%s", err, out)
		}
	}
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	return consumer
}

func initDepRepo(t *testing.T, workRoot, name string) string {
	t.Helper()
	dep := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, dep)
	writeFile(t, filepath.Join(dep, "go.mod"), "module "+depModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n")
	runGitIsolated(t, dep, "add", "go.mod", "dep.go")
	runGitIsolated(t, dep, "commit", "-m", "add go module")
	return dep
}

// readWorktreeGitdir reads the gitdir content from a linked worktree's owning
// main repo. For a worktree at wtPath, its .git file says
// "gitdir: <mainRepo>/.git/worktrees/<name>", so we read the file
// <mainRepo>/.git/worktrees/<name>/gitdir to get the registered path.
func readWorktreeGitdir(wtPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(wtPath, ".git"))
	if err != nil {
		return "", fmt.Errorf("read .git file: %w", err)
	}
	s := strings.TrimSpace(string(data))
	const prefix = "gitdir: "
	if !strings.HasPrefix(s, prefix) {
		return "", fmt.Errorf("unexpected .git file format in %s: %s", wtPath, s)
	}
	gitdir := strings.TrimSpace(s[len(prefix):])
	// gitdir file is at <mainRepo>/.git/worktrees/<name>/gitdir -> the content
	// is the path back to the worktree. We need to read that file.
	content, err := os.ReadFile(filepath.Join(gitdir, "gitdir"))
	if err != nil {
		return "", fmt.Errorf("read gitdir file at %s: %w", filepath.Join(gitdir, "gitdir"), err)
	}
	return strings.TrimSpace(string(content)), nil
}

// readWorktreeMainRepo reads a linked worktree's .git file and returns the main
// repository path it is registered under.
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
	mainRepo := filepath.Dir(filepath.Dir(filepath.Dir(gitdir)))
	if resolved, err := filepath.EvalSymlinks(mainRepo); err == nil {
		mainRepo = resolved
	}
	return mainRepo, nil
}

type propagateGoModJSON struct {
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path string `json:"Path"`
		} `json:"New"`
	} `json:"Replace"`
}

func propagateReadGoMod(modDir string) (*propagateGoModJSON, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var mod propagateGoModJSON
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func propagateReplacePathForModule(mod *propagateGoModJSON, modulePath string) string {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath {
			return repl.New.Path
		}
	}
	return ""
}
```
