# Scenario

**Bug**: wrk --status must count untracked (`??`) files as dirty `added`, not report clean

```
# clean committed checkout + one untracked file (not staged)
myrepo + ?? new.txt -> wrk --status -> dirty (1 added, 0 changed, 0 renamed, 0 deleted)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit `README.md` with subject `untracked dirty base` (repo is clean).
3. Create an untracked file `new.txt` (do not stage or commit).
4. Run `wrk --status` from the repo root.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, repo, "untracked dirty base")

	writeFile(t, filepath.Join(repo, "new.txt"), "untracked\n")

	req.RepoDir = repo
	req.MainRepo = repo
	return nil
}
```
