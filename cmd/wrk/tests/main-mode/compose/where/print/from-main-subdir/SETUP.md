# Scenario

**Feature**: wrk --main --where from a main-repo subdirectory prints main root (not subdir)

```
myrepo/pkg/tool -> wrk --main --where
  -> stdout myrepo abs path\n (not .../pkg/tool)
```

## Steps

1. Create main repo + nested subdir.
2. Install fake bash.
3. Run `wrk --main --where` with cwd = subdir.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, sub := initMainRepoSubdir(t, req, "pkg", "cmd", "tool")
	req.MainRepo = mainRepo
	req.RepoDir = sub
	installFakeBash(t, req, 0)
	setMainWhereArgs(req, "--main", "--where")
	return nil
}
```
