# Scenario

**Feature**: wrk --status reports deterministic added, changed, renamed, and deleted counts

```
# checkout has one porcelain entry in each status class
dirty myrepo -> wrk --status -> dirty (1 added, 1 changed, 1 renamed, 1 deleted)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit files used for changed, renamed, and deleted entries.
3. Modify one tracked file.
4. Stage one new file.
5. Rename one tracked file with `git mv`.
6. Delete one tracked file from the working tree.
7. Run `wrk --status` from the repo root.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, repo, "dirty base")

	writeFile(t, filepath.Join(repo, "rename-me.txt"), "old name\n")
	writeFile(t, filepath.Join(repo, "delete-me.txt"), "delete me\n")
	runGitIsolated(t, repo, "add", "rename-me.txt", "delete-me.txt")
	runGitIsolated(t, repo, "commit", "-m", "add dirty fixtures")

	writeFile(t, filepath.Join(repo, "README.md"), "# changed\n")
	writeFile(t, filepath.Join(repo, "added.txt"), "added\n")
	runGitIsolated(t, repo, "add", "added.txt")
	runGitIsolated(t, repo, "mv", "rename-me.txt", "renamed.txt")
	if err := os.Remove(filepath.Join(repo, "delete-me.txt")); err != nil {
		t.Fatalf("remove delete fixture: %v", err)
	}

	req.RepoDir = repo
	req.MainRepo = repo
	return nil
}
```
