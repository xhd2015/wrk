# Scenario

**Feature**: `WRK_SCAN_DEBUG=1` enables scan Debug without `-v`

```
# second run with env only
seed (quiet) → cache + projects
  -> WRK_SCAN_DEBUG=1 wrk --scan-git-repos ROOT   # no -v
  -> Debug=true from envTruthy
  -> stderr: scan: … mode=…
```

## Preconditions

- Parent seeded projects + product cache under FakeHome.
- **No** `-v` / `--verbose` on Args.
- Env truthy form under test: `1` (contract also allows `true` / `yes` case-insensitive; one form seals wiring).

## Steps

1. Set `ExtraEnv` to `WRK_SCAN_DEBUG=1` (last wins over ambient).
2. Keep Args as quiet `wrk --scan-git-repos <scan-root>` (no verbose flag).

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SCAN_DEBUG=1")
	req.Args = []string{"--scan-git-repos", scanRoot}
	return nil
}
```
