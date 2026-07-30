# Scenario

**Feature**: `wrk --scan-git-repos ~/Projects` shares home universe cache; stdout only under Projects

```
# cold seed full home (both mains → home/repos.json only; print-only)
wrk --scan-git-repos
  -> home/repos.json lists home-main + Projects/proj-main
  -> projects.json untouched (empty)

# filter root = ~/Projects (same product CacheRoot / home universe)
wrk --scan-git-repos $HOME/Projects
  -> still uses {CacheRoot}/home/* cache files
  -> stdout only abs(Projects/proj-main)
  -> home-main not on stdout; projects.json still empty
```

## Preconditions

- Parent fixtures + FakeHome isolation.
- Quiet full-home seed populates product **home** cache for both fixture mains.
- Scan never writes `projects.json` (print-only).

## Steps

1. Seed bare `--scan-git-repos` (no debug).
2. Assert home universe index exists (seed contract).
3. Set Args to `--scan-git-repos` + absolute `Projects` dir under FakeHome.
4. Force `WRK_SCAN_DEBUG=` empty.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	seedScanHomeNoDebug(t, req)

	indexPath := productHomeUniverseReposJSON(req.FakeHome)
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("seed must create home universe index %s: %v", indexPath, err)
	}

	projectsRoot := filepath.Join(req.FakeHome, "Projects")
	req.Args = []string{"--scan-git-repos", projectsRoot}
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SCAN_DEBUG=")
	return nil
}
```
