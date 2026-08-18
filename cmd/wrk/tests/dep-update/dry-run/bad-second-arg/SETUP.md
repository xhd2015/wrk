# Scenario

**Feature**: dry-run validates every dir arg first; a bad second arg prints no banner / no tree

```
valid first dep; second path missing
  -> wrk --dep-update <dep> <missing> --dry-run
  -> wrk: + no such dir
  -> no ==== banner; first dep not described as a half-plan
  -> go.mod unchanged
```

## Steps

1. Seed drop-replace-latest consumer + tagged dep.
2. Pass a missing second dir with `--dry-run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDropReplaceLatest(t, req)
	req.MissingPath = filepath.Join(req.WorkRoot, "does-not-exist-dep2")
	req.Args = []string{"--dep-update", req.DepDir, req.MissingPath, "--dry-run"}
	return nil
}
```
