# Scenario

**Feature**: multi-arg apply validates every dir first; missing second aborts with no writes

```
consumer + good dep; second path missing
  -> wrk --dep-replace <good> <missing>
  -> non-zero; no banner
  -> consumer go.mod unchanged (no partial replace)
```

## Steps

1. Seed single good dep.
2. Pass good dep then nonexistent path.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, true)
	req.MissingPath = filepath.Join(req.WorkRoot, "missing-second")
	req.Args = []string{"--dep-replace", req.DepDir, req.MissingPath}
	return nil
}
```
