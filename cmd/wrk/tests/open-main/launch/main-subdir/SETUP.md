# Scenario

**Feature**: wrk --main from a subdirectory of the main repo launches shell at main root

```
# cwd is nested inside main checkout
myrepo/pkg/cmd/tool (main) -> wrk --main
  -> fake shell cwd = myrepo (main root)
  -> exit 0; empty stdout; no install hint
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Create nested subdir `{WorkRoot}/myrepo/pkg/cmd/tool`.
3. Install fake bash (exit 0); set cwd to the nested subpath.
4. Run `wrk --main`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, sub := initMainRepoSubdir(t, req, "pkg", "cmd", "tool")
	req.RepoDir = sub
	req.MainRepo = mainRepo
	installFakeBash(t, req, 0)
	setMainArgs(req)
	return nil
}
```
