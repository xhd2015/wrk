# Scenario

**Feature**: `-v` on a home-subpath scan shows greppable `cache_base` and `filter`

```
# seed full home quietly
wrk --scan-git-repos

# second run: Projects root + verbose (Debug)
wrk --scan-git-repos -v $HOME/Projects
  -> stderr: scan: … and cache_base + filter tokens
  -> already-known after seed → empty stdout OK
```

## Preconditions

- Parent fixtures + FakeHome.
- Quiet full-home seed so cache is warm-eligible and projects already known.
- Second run under test: explicit Projects root with `-v`.

## Steps

1. Seed bare `--scan-git-repos` (no debug).
2. Set Args to `--scan-git-repos -v <Projects>`.
3. Force ambient `WRK_SCAN_DEBUG=` empty so only `-v` enables Debug.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	seedScanHomeNoDebug(t, req)

	projectsRoot := filepath.Join(req.FakeHome, "Projects")
	req.Args = []string{"--scan-git-repos", "-v", projectsRoot}
	// Last empty wins over ambient host WRK_SCAN_DEBUG for truthiness isolation.
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SCAN_DEBUG=")
	return nil
}
```
