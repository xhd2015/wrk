# Scenario

**Feature**: --color colors warning: prefix orange on ambiguous-cmd stderr line

```
# C3-wc: ambiguous cmd fixtures + --color
mod/ -> wrk --reinstall-local --dry-run --color
  -> stdout summary plain
  -> stderr: <orange #33>warning:</orange> bin foo: ambiguous under cmd (...); skipping
```

## Steps

1. Write two cmd package mains sharing bin `foo`.
2. Touch `$GOBIN/foo`.
3. Run with `--reinstall-local --dry-run --color`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-warn-color")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "nested", "foo"))
	touchBin(t, req.BinDir, "foo")
	req.Args = []string{"--reinstall-local", "--dry-run", "--color"}
	return nil
}
```
