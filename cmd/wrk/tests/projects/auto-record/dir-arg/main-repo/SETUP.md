# Scenario

**Feature**: auto-record via `wrk <mainRepo>`

```
WorkRoot -> wrk myrepo --list -> projects.json records myrepo main path
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk <mainRepo> --list` from `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.TargetDir = mainRepo
	return nil
}
```