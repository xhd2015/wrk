# Scenario

**Feature**: wrk --cd missing absolute path errors

```
wrk --cd /WorkRoot/does-not-exist -> non-zero; does not exist
```

## Steps

1. Use a guaranteed-missing absolute path under WorkRoot.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	missing := filepath.Join(req.WorkRoot, "does-not-exist")
	req.MainRepo = missing
	setCDFlagThenPath(req, missing)
	return nil
}
```
