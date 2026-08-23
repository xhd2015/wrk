# Scenario

**Bug**: wrk --status on a nested independent checkout with only untracked files reports dirty untracked, while a clean root stays clean

```
# root clean; nested independent git tools/child has only ?? untracked
myrepo (clean) + myrepo/tools/child + ?? new.txt -> wrk --status
  -> Dir "." clean; Dir "tools/child" dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit a root `.gitignore` containing `tools/` so the nested independent checkout is not
   counted as untracked on the parent (parent porcelain stays clean when untracked is included).
3. Initialize `{WorkRoot}/myrepo/tools/child` as an independent git repo on branch `main`.
4. Create an untracked file `new.txt` under `tools/child` (do not stage or commit).
5. Run `wrk --status` from `{WorkRoot}/myrepo`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "myrepo")
	child := filepath.Join(repo, "tools", "child")

	statusInitRepoWithSubject(t, repo, "root status repo")
	// Nested independent git dirs appear as ?? on the parent when untracked files are
	// included. Ignore tools/ so the root block stays clean while the child is still
	// discovered by scan_repo and reports its own status.
	writeFile(t, filepath.Join(repo, ".gitignore"), "tools/\n")
	runGitIsolated(t, repo, "add", ".gitignore")
	runGitIsolated(t, repo, "commit", "-m", "ignore nested tools")

	statusInitRepoWithSubject(t, child, "child status repo")
	writeFile(t, filepath.Join(child, "new.txt"), "untracked\n")

	req.RepoDir = repo
	req.MainRepo = repo
	req.DepPath = child
	return nil
}
```
