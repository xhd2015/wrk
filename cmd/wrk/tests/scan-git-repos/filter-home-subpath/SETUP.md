# Scenario

**Feature**: wrk --scan-git-repos maps home roots to universe home; subpaths share home cache and filter emit

```
# FakeHome sandbox (product CacheRoot under $HOME/.cache/git-repo-scan)
HOME=FakeHome (= WorkRoot)
  FakeHome/home-main          # main outside Projects
  FakeHome/Projects/proj-main # main under ~/Projects

# P5 two-base + filter (home cases only in this branch)
wrk --scan-git-repos                  # default root ~
  -> universe=home
  -> product cache files under $HOME/.cache/git-repo-scan/home/
  -> emit/record mains under ~

wrk --scan-git-repos ~/Projects
  -> same home universe cache files (not a separate Projects universe)
  -> filter emit to paths under Projects only
  -> stdout only newly recorded mains under Projects

wrk --scan-git-repos ~/Projects -v
  -> stderr greppable cache_base + filter (with scan: phase lines)
```

## Preconditions

- **FakeHome = WorkRoot** so product default CacheRoot is
  `{FakeHome}/.cache/git-repo-scan` only (never the real user cache).
- Fixture always has two mains under FakeHome:
  - `{FakeHome}/home-main` — under home, **outside** Projects
  - `{FakeHome}/Projects/proj-main` — under `~/Projects`
- Cwd remains non-git `{WorkRoot}` so auto-record does not pollute projects.json.
- Outside-home / root-universe paths are out of scope here (hard under FakeHome-only
  isolation; product may still map them later).

## Steps

1. Set `req.FakeHome = req.WorkRoot`.
2. Create `home-main` and `Projects/proj-main` as main git repos.
3. Stash paths: `MainRepo` = proj-main, `SecondRepo` = home-main.
4. Register product-cache path helpers for descendants.
5. Leaves set Args (and optional seed / clear projects) for the case under test.

## Context

- Product cache layout (library P1–P4): `{CacheRoot}/home/repos.json` is universe
  **home**; mirror marks live under `{CacheRoot}/mirror/...`.
- Sharing home cache: scanning `~/Projects` must keep using universe **home** files
  under the same product CacheRoot — not invent a separate universe keyed only by
  the Projects path.
- Emit filter: stdout and new `source=scan` records for a Projects root cover only
  mains under Projects; `home-main` must not appear on that run's stdout.
- Debug (P5): when `-v` / Debug is on, stderr should expose greppable **`cache_base`**
  and **`filter`** tokens (alongside existing `scan:` / `mode=` lines).

```go
import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

// productGitRepoScanCacheRoot is the product default when CacheRoot is empty:
// $HOME/.cache/git-repo-scan under FakeHome.
func productGitRepoScanCacheRoot(fakeHome string) string {
	return filepath.Join(fakeHome, ".cache", "git-repo-scan")
}

func productHomeUniverseReposJSON(fakeHome string) string {
	return filepath.Join(productGitRepoScanCacheRoot(fakeHome), "home", "repos.json")
}

func productRootUniverseReposJSON(fakeHome string) string {
	return filepath.Join(productGitRepoScanCacheRoot(fakeHome), "root", "repos.json")
}

// seedScanHomeNoDebug runs quiet bare --scan-git-repos with FakeHome so both
// fixture mains are discovered into projects.json and the product home universe.
func seedScanHomeNoDebug(t *testing.T, req *Request) {
	t.Helper()
	if req.FakeHome == "" {
		t.Fatal("seedScanHomeNoDebug requires FakeHome")
	}
	bin := getWrkBin(t)
	cmd := exec.Command(bin, "--scan-git-repos")
	cmd.Dir = req.RepoDir
	env := wrkEnv(req)
	env = append(env, "WRK_SCAN_DEBUG=")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("seed bare --scan-git-repos: exit %d output=%q", ee.ExitCode(), string(out))
		}
		t.Fatalf("seed bare --scan-git-repos: %v output=%q", err, string(out))
	}
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")
	assertScanProjectRecorded(t, req.WrkHome, req.SecondRepo, "scan")
}

// clearScanProjectsJSON removes projects.json so a later scan can re-record
// and prove emit filter via stdout without treating paths as already-known.
func clearScanProjectsJSON(t *testing.T, wrkHome string) {
	t.Helper()
	if err := os.Remove(scanProjectsJSONPath(wrkHome)); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove projects.json: %v", err)
	}
}

// homeUniverseIndexHasPath reports whether home/repos.json lists path (best-effort;
// missing file → false).
func homeUniverseIndexHasPath(t *testing.T, fakeHome, path string) bool {
	t.Helper()
	data, err := os.ReadFile(productHomeUniverseReposJSON(fakeHome))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		t.Fatalf("read home/repos.json: %v", err)
	}
	var doc struct {
		Universe string `json:"universe"`
		Repos    []struct {
			Path string `json:"path"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse home/repos.json: %v", err)
	}
	want := resolveScanPath(t, path)
	for _, r := range doc.Repos {
		if resolveScanPath(t, r.Path) == want {
			return true
		}
	}
	return false
}

func Setup(t *testing.T, req *Request) error {
	// Isolate HOME → product default cache only under sandbox.
	req.FakeHome = req.WorkRoot

	// home-main: under ~ but outside Projects
	homeMain := initScanMainRepo(t, req.WorkRoot, "home-main")
	req.SecondRepo = homeMain

	// Projects/proj-main: classic subpath filter target
	projectsDir := filepath.Join(req.WorkRoot, "Projects")
	mkdirAll(t, projectsDir)
	projMain := initScanMainRepo(t, projectsDir, "proj-main")
	req.MainRepo = projMain

	// Keep helpers referenced for leaf packages.
	_ = productGitRepoScanCacheRoot
	_ = productHomeUniverseReposJSON
	_ = productRootUniverseReposJSON
	_ = seedScanHomeNoDebug
	_ = clearScanProjectsJSON
	_ = homeUniverseIndexHasPath
	return nil
}
```
