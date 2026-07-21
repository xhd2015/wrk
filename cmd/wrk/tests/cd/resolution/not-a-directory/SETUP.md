# Scenario

**Feature**: wrk --cd rejects a path that exists as a regular file

```
workspace/notdir is a file
wrk --cd workspace/notdir (abs) -> non-zero; not a directory (or similar)
```

## Steps

1. Create a regular file under WorkRoot.
2. `wrk --cd <file-abs>`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	filePath := filepath.Join(req.WorkRoot, "notdir")
	writeFile(t, filePath, "not a directory\n")
	req.MainRepo = resolvePath(t, filePath)
	setCDFlagThenPath(req, req.MainRepo)
	return nil
}
```
