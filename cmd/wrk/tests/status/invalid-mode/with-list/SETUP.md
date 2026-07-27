# Scenario

**Feature**: wrk --status rejects --list in the same invocation

```
# both flags are standalone reporting modes
wrk --status --list -> error (mutually exclusive)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Run `wrk --status --list` from the repo root.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, repo, "status mode conflict")

	req.RepoDir = repo
	req.Args = []string{"--status", "--list"}
	return nil
}
```
