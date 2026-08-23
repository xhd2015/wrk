# Scenario

**Feature**: staged-new then further edited (`AM`) counts once as added

```
# AM porcelain line → added only (not also changed)
myrepo + AM edit.txt -> wrk --status -> dirty (1 staged, 0 changed, 0 renamed, 0 deleted, 0 untracked)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit `README.md` with subject `status format am base`.
3. Create and stage `edit.txt`, then modify it again without restaging.
4. Run `wrk --status` from the repo root.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusFormatInitRepo(t, repo, "status format am base")
	writeFile(t, filepath.Join(repo, "edit.txt"), "v1\n")
	runGitIsolated(t, repo, "add", "edit.txt")
	writeFile(t, filepath.Join(repo, "edit.txt"), "v2\n")
	req.RepoDir = repo
	req.MainRepo = repo
	return nil
}
```
