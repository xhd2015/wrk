# Scenario

**Feature**: successful wrk --main records events.jsonl command "main"

```
myrepo/pkg/tool -> wrk --main
  -> exit 0
  -> events.jsonl last: command=main, exit_code=0, args include --main
```

## Steps

1. Create main repo + nested subdir.
2. Install fake bash (exit 0).
3. Run `wrk --main`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, sub := initMainRepoSubdir(t, req, "pkg", "tool")
	req.MainRepo = mainRepo
	req.RepoDir = sub
	installFakeBash(t, req, 0)
	setMainArgs(req)
	return nil
}
```
