# Scenario

**Feature**: wrk --cd missing absolute path errors

```
wrk --cd /WorkRoot/does-not-exist -> non-zero; does not exist
```

## Steps

1. Use a guaranteed-missing absolute path under WorkRoot.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	missing := filepath.Join(req.WorkRoot, "does-not-exist")
	req.MainRepo = missing
	setCDFlagThenPath(req, missing)
	return nil
}
```
