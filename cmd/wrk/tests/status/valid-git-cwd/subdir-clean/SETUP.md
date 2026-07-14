# Scenario

**Feature**: wrk --status invoked from a subdirectory still reports root-relative dot

```
# process cwd is below the checkout top
myrepo/sub/dir -> wrk --status -> Dir "."
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Create and commit `sub/dir/file.txt`.
3. Run `wrk --status` from `{WorkRoot}/myrepo/sub/dir`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, repo, "initial status subdir")

	nested := filepath.Join(repo, "sub", "dir")
	mkdirAll(t, nested)
	writeFile(t, filepath.Join(nested, "file.txt"), "subdir\n")
	runGitIsolated(t, repo, "add", "sub/dir/file.txt")
	runGitIsolated(t, repo, "commit", "-m", "add subdir file")

	req.RepoDir = nested
	req.MainRepo = repo
	return nil
}
```
