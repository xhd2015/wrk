# Scenario

**Feature**: --dep-update fails when dep dir is not a go module

```
consumer present; plain dir without go.mod
  -> wrk --dep-update <plain>
  -> non-zero; go.mod unchanged
```

## Steps

1. Seed consumer with replace.
2. Create non-module directory and pass it as dep.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDropReplaceLatest(t, req)
	plain := filepath.Join(req.WorkRoot, "plain-not-mod")
	mkdirAll(t, plain)
	writeFile(t, filepath.Join(plain, "readme.txt"), "not a module\n")
	plain = resolvePath(t, plain)
	req.Args = []string{"--dep-update", plain}
	return nil
}
```
