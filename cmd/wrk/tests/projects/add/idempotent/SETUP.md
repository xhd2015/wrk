# Scenario

**Feature**: duplicate auto + manual add is idempotent

```
wrk --list (auto) then wrk --add <same main> -> single entry, source stays auto
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --list` to auto-record.
3. Run `wrk --add <mainRepo>`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	runWrkWithArgs(t, req, mainRepo, "--list")
	req.MainRepo = mainRepo
	req.Args = []string{"--add", mainRepo}
	return nil
}
```