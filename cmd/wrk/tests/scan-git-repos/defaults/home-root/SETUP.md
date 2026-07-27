# Scenario

**Feature**: bare `wrk --scan-git-repos` scans `$HOME` and records a main under home

```
# HOME = WorkRoot; main at $HOME/repo-a; no $HOME/Projects
HOME=WorkRoot
  -> wrk --scan-git-repos   (no ROOT args)
  -> discovers abs(repo-a)
  -> projects.json unchanged
  -> stdout absolute main path
```

## Preconditions

- `FakeHome = WorkRoot` so product `os.UserHomeDir()` resolves to the isolated temp home.
- Main git repo at `{WorkRoot}/repo-a` (directly under home).
- **No** `{WorkRoot}/Projects` directory — product default root is `$HOME` (`~`), not `~/Projects`; leaving Projects absent ensures a Projects-only default cannot false-green.

## Steps

1. Set `req.FakeHome = req.WorkRoot`.
2. Create main repo `repo-a` under WorkRoot (home).
3. Assert Projects dir is absent.
4. Set Args to bare `--scan-git-repos` (no roots).

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Isolate HOME so bare --scan-git-repos defaults into this WorkRoot.
	req.FakeHome = req.WorkRoot

	// One main under home; no explicit ROOT CLI args.
	mainRepo := initScanMainRepo(t, req.WorkRoot, "repo-a")
	req.MainRepo = mainRepo

	// Guard: Projects must not exist or a Projects-only default could false-green.
	projectsDir := filepath.Join(req.WorkRoot, "Projects")
	if st, err := os.Stat(projectsDir); err == nil && st.IsDir() {
		t.Fatalf("fixture must not create Projects dir (masks home-default bug): %s", projectsDir)
	}

	req.Args = []string{"--scan-git-repos"}
	// RepoDir stays WorkRoot (parent): WorkRoot itself is not a git repo.
	return nil
}
```
