# Scenario

**Feature**: bare --install is a parse error (at least one arg required)

```
# I10: ./cmd/tool exists but no name is given
mod/ -> wrk --install
  -> non-zero; library wording: --install requires a value
  -> stdout empty (never reaches planning)
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-noarg`.
2. Write `./cmd/tool` as `package main` (available, but not requested).
3. Run bare `wrk --install`; expect the lessflags minimum-arg error.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-noarg")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	req.Args = []string{"--install"}
	return nil
}
```
