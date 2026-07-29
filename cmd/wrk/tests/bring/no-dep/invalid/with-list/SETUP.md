# Scenario

**Feature**: --no-dep with --list is rejected

```
wrk --list --no-dep -> non-zero
  stderr: --no-dep is only valid with --dep, --bring, or --all-deps
```

## Steps

1. Create a main repo so `--list` would otherwise be valid.
2. Run `wrk --list --no-dep` from that repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	repo := filepath.Join(req.WorkRoot, "list-no-dep")
	initGitRepoOnMain(t, repo)
	req.RepoDir = repo
	req.Args = []string{"--list", "--no-dep"}
	return nil
}
```
