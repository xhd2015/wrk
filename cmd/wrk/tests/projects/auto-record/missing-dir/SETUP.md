# Scenario

**Feature**: no auto-record when `<dir>` does not exist

```
WorkRoot -> wrk <nonexistent> --list -> error; no projects.json created
```

## Steps

1. Run `wrk <missingPath> --list` from `{WorkRoot}`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.TargetDir = filepath.Join(req.WorkRoot, "does-not-exist")
	req.RepoDir = req.WorkRoot
	return nil
}
```