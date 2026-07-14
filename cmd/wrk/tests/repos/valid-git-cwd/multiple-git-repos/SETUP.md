# Scenario

**Feature**: wrk --repos reports every git directory discovered under the checkout root

```
myrepo + myrepo/tools/child -> wrk --repos -> "." and "tools/child"
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Initialize `{WorkRoot}/myrepo/tools/child` as an independent git repo on branch `main`.
3. Run `wrk --repos` from `{WorkRoot}/myrepo`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	child := filepath.Join(repo, "tools", "child")

	reposInitRepo(t, repo)
	reposInitRepo(t, child)

	req.RepoDir = repo
	return nil
}
```
