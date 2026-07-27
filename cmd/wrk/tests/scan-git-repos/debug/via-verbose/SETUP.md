# Scenario

**Feature**: `-v` / `--verbose` enables scan Debug on a second warm-ish scan

```
# second run with -v (no WRK_SCAN_DEBUG)
seed (quiet) → cache + projects
  -> wrk --scan-git-repos -v ROOT
  -> Debug=true from verbose
  -> stderr: scan: … mode=warm|cold …
```

## Preconditions

- Parent seeded projects + product cache under FakeHome.
- `WRK_SCAN_DEBUG` forced empty so only `-v` turns Debug on.
- Explicit ROOT (scan-root), not bare home default.

## Steps

1. Force `WRK_SCAN_DEBUG=` via ExtraEnv (ambient isolation).
2. Set Args to `wrk --scan-git-repos -v <scan-root>` (second run).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	forceScanDebugEnvOff(req)
	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	// -v after mode flag is fine; both -v and --verbose are accepted globally.
	req.Args = []string{"--scan-git-repos", "-v", scanRoot}
	return nil
}
```
