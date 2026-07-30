# Scenario

**Feature**: `-v` on a home-subpath scan shows greppable `cache_base` and `filter`

```
# seed full home quietly (warms home/repos.json only; print-only)
wrk --scan-git-repos

# second run: Projects root + verbose (Debug)
wrk --scan-git-repos -v $HOME/Projects
  -> stderr: scan: … and cache_base + filter tokens
  -> always-print: Projects main still on stdout once
```

## Preconditions

- Parent fixtures + FakeHome.
- Quiet full-home seed so product home universe cache is warm-eligible.
- Second run under test: explicit Projects root with `-v`.

## Steps

1. Seed bare `--scan-git-repos` (no debug; no projects.json write).
2. Set Args to `--scan-git-repos -v <Projects>`.
3. Force ambient `WRK_SCAN_DEBUG=` empty so only `-v` enables Debug.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	seedScanHomeNoDebug(t, req)

	projectsRoot := filepath.Join(req.FakeHome, "Projects")
	req.Args = []string{"--scan-git-repos", "-v", projectsRoot}
	// Last empty wins over ambient host WRK_SCAN_DEBUG for truthiness isolation.
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SCAN_DEBUG=")
	return nil
}
```
