# Scenario

**Feature**: `wrk --push --dry-run` plans branch push without remote mutation

```
# main + origin; origin tip snapshotted
myrepo (main) + origin
  -> wrk --push --dry-run
  -> would: git push origin main
  -> origin/main unchanged; no confirm "pushed …" line
```

## Steps

1. Seed main with bare origin (upstream set).
2. Capture origin/main SHA before run.
3. Run `wrk --push --dry-run`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushMainWithOrigin(t, req)
	// Snapshot origin tip via WorkRoot file for Assert.
	sha := revParseRef(t, req.OriginBare, "refs/heads/main")
	writeFile(t, filepath.Join(req.WorkRoot, "origin-main-before"), sha+"\n")
	req.Args = []string{"--push", "--dry-run"}
	return nil
}
```
