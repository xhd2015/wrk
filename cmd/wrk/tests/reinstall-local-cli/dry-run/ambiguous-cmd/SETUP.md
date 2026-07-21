# Scenario

**Feature**: dry-run omits ambiguous cmd bin and prints warning on stderr

```
# C3-amb: ./cmd/foo + ./cmd/nested/foo + GOBIN/foo (no script)
mod/ -> wrk --reinstall-local --dry-run
  -> stdout: would: reinstall 0 binaries (0 skipped)
  -> stderr: warning: bin foo: ambiguous under cmd (./cmd/foo, ./cmd/nested/foo); skipping
  # no would: go install for foo
```

## Steps

1. Write `go.mod` with module `example.com/cli-amb-cmd`.
2. Write `./cmd/foo` and `./cmd/nested/foo` as `package main`.
3. Touch `$GOBIN/foo`.
4. Run `wrk --reinstall-local --dry-run` from module root.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-amb-cmd")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "nested", "foo"))
	touchBin(t, req.BinDir, "foo")
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
