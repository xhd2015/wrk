# Scenario

**Feature**: --main --reinstall-local --dry-run records --main in events.jsonl args (H2)

```
# H2: wrk --main --reinstall-local --dry-run success -> events.jsonl last event
git-mod/ -> wrk --main --reinstall-local --dry-run
  -> exit 0
  -> last event: command=reinstall-local, exit_code=0
  -> args include --main, --reinstall-local, and --dry-run
```

## Steps

1. Init a git repo with a single go.mod module (required for useMain scan root).
2. Write `./cmd/missing` as `package main` (skip-only plan; no install work).
3. Commit so ShowToplevel / ResolveMainRepo succeed.
4. Run `wrk --main --reinstall-local --dry-run` from that repo root.
5. Assert last `events.jsonl` event (do not re-invoke wrk before read).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepoOnMain(t, repo)
	writeGoMod(t, repo, "example.com/cli-events-main-dry")
	writePackageMain(t, filepath.Join(repo, "cmd", "missing"))
	gitCommitAll(t, repo, "init events main-dry-run fixture")
	// skip-only plan: no GOBIN stub — keeps dry-run fast and exit 0
	req.ModuleRoot = resolvePath(t, repo)
	req.Args = []string{"--main", "--reinstall-local", "--dry-run"}
	return nil
}
```
