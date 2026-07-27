# Scenario

**Feature**: wrk --repos shows the root checkout as dot when run from the root

```
myrepo -> wrk --repos -> "."
```

## Steps

1. Initialize `{WorkRoot}` as a git repo on branch `main`.
2. Run `wrk --repos` from `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	reposInitRepo(t, req.WorkRoot)
	req.RepoDir = req.WorkRoot
	return nil
}
```
