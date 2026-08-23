# Scenario

**Feature**: wrk --status counts staged new files as `staged`

```
# clean checkout + staged new file (A)
myrepo + A new.txt -> wrk --status -> dirty (1 staged, 0 changed, 0 renamed, 0 deleted, 0 untracked)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit `README.md` with subject `status format staged added base`.
3. Create and stage `new.txt` (do not commit).
4. Run `wrk --status` from the repo root.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusFormatInitRepo(t, repo, "status format staged added base")
	writeFile(t, filepath.Join(repo, "new.txt"), "staged\n")
	runGitIsolated(t, repo, "add", "new.txt")
	req.RepoDir = repo
	req.MainRepo = repo
	return nil
}
```
