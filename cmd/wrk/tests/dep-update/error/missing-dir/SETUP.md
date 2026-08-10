# Scenario

**Feature**: --dep-update fails when dep directory does not exist

```
consumer present; dep path missing
  -> wrk --dep-update <missing>
  -> non-zero; go.mod unchanged
```

## Steps

1. Seed consumer+tagged dep (consumer only needed for baseline).
2. Pass nonexistent path.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDropReplaceLatest(t, req)
	req.MissingPath = filepath.Join(req.WorkRoot, "does-not-exist-dep")
	req.Args = []string{"--dep-update", req.MissingPath}
	return nil
}
```
