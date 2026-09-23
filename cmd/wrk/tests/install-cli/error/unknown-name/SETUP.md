# Scenario

**Feature**: a name with no discovered candidate is a hard error

```
# I11: only ./cmd/known exists; ask for nope
mod/ -> wrk --install nope
  -> non-zero; stderr: wrk: --install: no install candidate for "nope"
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-unknown`.
2. Write `./cmd/known` as `package main`.
3. Run `wrk --install nope`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-unknown")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "known"))
	req.Args = []string{"--install", "nope"}
	return nil
}
```
