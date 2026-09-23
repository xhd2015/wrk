# Scenario

**Feature**: two cmd dirs with the same basename make the name ambiguous

```
# I12: ./cmd/dup + ./cmd/nested/dup (both package main)
mod/ -> wrk --install dup
  -> non-zero; stderr: bin "dup" is ambiguous (./cmd/dup, ./cmd/nested/dup)
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-ambiguous`.
2. Write `./cmd/dup` and `./cmd/nested/dup` as `package main`.
3. Run `wrk --install dup`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-ambiguous")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "dup"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "nested", "dup"))
	req.Args = []string{"--install", "dup"}
	return nil
}
```
