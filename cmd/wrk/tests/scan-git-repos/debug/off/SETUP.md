# Scenario

**Feature**: without `-v` and without truthy `WRK_SCAN_DEBUG`, scan path stays quiet (no `scan:`)

```
# second run, debug off
seed (quiet) → cache + projects
  -> wrk --scan-git-repos ROOT   # no -v, WRK_SCAN_DEBUG empty/non-truthy
  -> Debug=false
  -> stderr has zero scan: markers
```

## Preconditions

- Parent seeded projects + product cache (non-vacuous: a Debug-on Scan would emit `scan:`).
- No verbose flag; force `WRK_SCAN_DEBUG=` so ambient host env cannot false-green or false-red.

## Steps

1. Force `WRK_SCAN_DEBUG=` via ExtraEnv.
2. Keep Args as quiet `wrk --scan-git-repos <scan-root>`.

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
	req.Args = []string{"--scan-git-repos", scanRoot}
	return nil
}
```
