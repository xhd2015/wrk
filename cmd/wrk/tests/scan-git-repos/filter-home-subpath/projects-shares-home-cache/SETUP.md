# Scenario

**Feature**: `wrk --scan-git-repos ~/Projects` shares home universe cache; stdout only under Projects

```
# cold seed full home (both mains → home/repos.json + projects.json)
wrk --scan-git-repos
  -> home/repos.json lists home-main + Projects/proj-main

# clear projects so second run can re-record via emit filter
rm projects.json

# filter root = ~/Projects (same product CacheRoot / home universe)
wrk --scan-git-repos $HOME/Projects
  -> still uses {CacheRoot}/home/* cache files
  -> stdout only abs(Projects/proj-main)
  -> home-main not on stdout; not re-recorded this run
```

## Preconditions

- Parent fixtures + FakeHome isolation.
- Quiet full-home seed populates product **home** cache and both projects.
- `projects.json` cleared after seed so stdout proves emit filter (paths not
  already-known).

## Steps

1. Seed bare `--scan-git-repos` (no debug).
2. Assert home universe index exists (seed contract).
3. Clear `projects.json`.
4. Set Args to `--scan-git-repos` + absolute `Projects` dir under FakeHome.
5. Force `WRK_SCAN_DEBUG=` empty.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	seedScanHomeNoDebug(t, req)

	indexPath := productHomeUniverseReposJSON(req.FakeHome)
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("seed must create home universe index %s: %v", indexPath, err)
	}

	clearScanProjectsJSON(t, req.WrkHome)

	projectsRoot := filepath.Join(req.FakeHome, "Projects")
	req.Args = []string{"--scan-git-repos", projectsRoot}
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SCAN_DEBUG=")
	return nil
}
```
