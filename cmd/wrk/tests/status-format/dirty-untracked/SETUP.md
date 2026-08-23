# Scenario

**Feature**: wrk --status prints dirty wording with untracked as `untracked`

```
# clean committed checkout + one untracked file (not staged)
myrepo + ?? new.txt -> wrk --status -> dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit `README.md` with subject `status format dirty base` (repo is clean).
3. Create an untracked file `new.txt` (do not stage or commit).
4. Run `wrk --status` from the repo root.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusFormatInitRepo(t, repo, "status format dirty base")
	writeFile(t, filepath.Join(repo, "new.txt"), "untracked\n")
	req.RepoDir = repo
	req.MainRepo = repo
	return nil
}
```
