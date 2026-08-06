# Scenario

**Feature**: `wrk --push -f --dry-run` prints force-with-lease plan line

```
myrepo + origin
  -> wrk --push -f --dry-run
  -> would: git push --force-with-lease origin main
  -> origin/main unchanged
```

## Steps

1. Seed main with bare origin (upstream set).
2. Snapshot origin/main SHA.
3. Run `wrk --push -f --dry-run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushMainWithOrigin(t, req)
	sha := revParseRef(t, req.OriginBare, "refs/heads/main")
	writeFile(t, filepath.Join(req.WorkRoot, "origin-main-before"), sha+"\n")
	req.Args = []string{"--push", "-f", "--dry-run"}
	return nil
}
```
