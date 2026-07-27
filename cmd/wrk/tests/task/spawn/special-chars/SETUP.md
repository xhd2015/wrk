# Scenario

**Feature**: --task with capitals, symbols, and unicode produces a clean slug

```
wrk --task "Fix: Login & Signup!!! (urgent)" -> slug = "fix-login-signup-urgent"
```

## Steps

1. Create repo.
2. Set req.TaskDesc so Run injects --task.
3. Verify slug is sanitized: lowercase, non-alphanumeric → "-", collapsed, trimmed.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	req.TaskDesc = "Fix: Login & Signup!!! (urgent)"
	req.RepoDir = mainRepo
	expectedSlug := slugify(req.TaskDesc)
	req.WtBranch = branchNameWithTask("main", wrkDate, expectedSlug, 0)
	return nil
}
```