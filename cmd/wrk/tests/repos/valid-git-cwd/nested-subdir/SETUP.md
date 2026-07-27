# Scenario

**Feature**: wrk --repos invoked from a subdirectory still reports root-relative dot

```
myrepo/sub/dir -> wrk --repos -> "."
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Create `{WorkRoot}/myrepo/sub/dir`.
3. Run `wrk --repos` from `{WorkRoot}/myrepo/sub/dir`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "myrepo")
	reposInitRepo(t, repo)
	subdir := filepath.Join(repo, "sub", "dir")
	mkdirAll(t, subdir)

	req.RepoDir = subdir
	return nil
}
```
