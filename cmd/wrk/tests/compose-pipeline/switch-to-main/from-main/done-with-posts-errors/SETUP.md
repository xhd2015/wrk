# Scenario

**Feature**: `--done` with post stages on main still hard-gates before pipeline (linked worktree)

```
myrepo (main) -> wrk --done --sync --tag-next --push -y
  -> non-zero before running posts
  -> stderr names --done + linked worktree (not a silent tag-on-main path)
```

## Steps

1. Main repo with go.mod.
2. Run done + posts from main.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\n\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--sync", "--tag-next", "--push", "-y"}
	return nil
}
```
