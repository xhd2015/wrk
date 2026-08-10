# Scenario

**Feature**: multi-arg fail-fast stops on second missing dir; first replace may remain (D3)

```
consumer + good dep; second path missing
  -> wrk --dep-replace <good> <missing>
  -> non-zero
  -> first dep absolute replace present
  -> no replace for second
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
