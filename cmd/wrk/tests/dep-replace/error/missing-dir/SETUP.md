# Scenario

**Feature**: --dep-replace fails when dep directory does not exist

```
consumer go.mod present; dep path missing
  -> wrk --dep-replace <missing>
  -> non-zero; go.mod unchanged
```

## Steps

1. Seed consumer+dep fixtures (consumer only needed).
2. Pass a nonexistent path as the dep arg.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, true)
	req.MissingPath = filepath.Join(req.WorkRoot, "does-not-exist-dep")
	req.Args = []string{"--dep-replace", req.MissingPath}
	return nil
}
```
