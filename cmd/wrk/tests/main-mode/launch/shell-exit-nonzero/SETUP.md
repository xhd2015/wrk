# Scenario

**Feature**: wrk --main propagates non-zero interactive shell exit code

```
# fake bash exits 42; cwd is main subdir (not already-at-root)
myrepo/pkg/tool -> wrk --main -> wrk exit code 42
```

## Steps

1. Create main repo + nested subdir.
2. Install fake bash with exit 42.
3. Run `wrk --main` from the subdir.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, sub := initMainRepoSubdir(t, req, "pkg", "tool")
	req.MainRepo = mainRepo
	req.RepoDir = sub
	installFakeBash(t, req, 42)
	setMainArgs(req)
	return nil
}
```
