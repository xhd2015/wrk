# Scenario

**Feature**: missing registry path soft-warns on stderr; remaining projects print

```
projects.json = [does-not-exist, good]
good: example.com/good (dir=.)
wrk --projects-dep-graph
  -> stderr: warning: … missing path …
  -> exit 0
  -> stdout: good project block + footer 1 project · 1 module · 0 cross-edges
```

## Steps

1. Create one good project with module `example.com/good`.
2. Seed projects.json with a non-existent path **and** the good path.
3. Expect warning on stderr mentioning the missing path; graph includes only good.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	goodPath := filepath.Join(req.WorkRoot, "repos", "good")
	initSingleModuleRepo(t, goodPath, "example.com/good")
	goodPath = resolvePath(t, goodPath)
	req.GoodPath = goodPath

	missingPath := filepath.Join(req.WorkRoot, "does-not-exist", "gone")
	// Do not create missingPath on disk.
	req.MissingPath = missingPath
	writeProjectsJSON(t, req.WrkHome, missingPath, goodPath)
	return nil
}
```
