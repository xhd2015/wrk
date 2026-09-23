# Scenario

**Feature**: an install failure is soft (exit 0 + stderr warning)

```
# I7: ./cmd/broken does not compile
mod/ -> wrk --install broken
  -> go install ./cmd/broken (compile error on stderr)
  -> installed 0, failed 1; exit 0; stderr warning: install finished with 1 failed
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-broken`.
2. Write `./cmd/broken` as `package main` that references an undefined symbol.
3. Run `wrk --install broken`; expect exit 0 and the soft-failure warning.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-broken")
	writeBrokenPackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "broken"))
	req.Args = []string{"--install", "broken"}
	return nil
}
```
