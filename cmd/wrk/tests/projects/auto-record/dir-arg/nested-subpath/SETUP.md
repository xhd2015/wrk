# Scenario

**Feature**: auto-record via `wrk <nestedSubpath>`

```
WorkRoot -> wrk myrepo/pkg/nested --list -> projects.json records myrepo main path
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Create nested subdir `{WorkRoot}/myrepo/pkg/nested`.
3. Run `wrk <nestedSubpath> --list` from `{WorkRoot}`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	subpath := filepath.Join(mainRepo, "pkg", "nested")
	mkdirAll(t, subpath)
	req.MainRepo = mainRepo
	req.TargetDir = subpath
	return nil
}
```