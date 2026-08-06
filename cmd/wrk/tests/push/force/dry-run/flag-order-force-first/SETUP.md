# Scenario

**Feature**: flag order free — `--force --push --dry-run` same plan as push-first

```
myrepo + origin
  -> wrk --force --push --dry-run
  -> would: git push --force-with-lease origin main
  -> origin/main unchanged
```

## Steps

1. Seed main with bare origin.
2. Snapshot origin/main SHA.
3. Run `wrk --force --push --dry-run` (force before push).

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
	req.Args = []string{"--force", "--push", "--dry-run"}
	return nil
}
```
