# Scenario

**Feature**: wrk wires `scan_repo.Options.Debug` (and Stderr) when `-v`/`--verbose` OR `WRK_SCAN_DEBUG` is truthy

```
# P3 wiring (CLI → library)
wrk --scan-git-repos ROOT
  -> debug := verbose || envTruthy(WRK_SCAN_DEBUG)
     # truthy: 1, true, yes (case-insensitive)
  -> scan_repo.Scan(Options{Debug: debug, Stderr: os.Stderr, …})
  -> when debug: stderr has greppable scan: phase lines (mode=warm|cold, …)
  -> when !debug: zero scan: markers from the scan path

# product default cache under FakeHome
HOME=FakeHome
  -> CacheRoot empty → $HOME/.cache/git-repo-scan (sandboxed)
```

## Preconditions

- **FakeHome = WorkRoot** so product default `CacheRoot` (`$HOME/.cache/git-repo-scan`) lands only under the isolated sandbox home — never the real user cache.
- Explicit small scan root `{WorkRoot}/scan-root` with one main `myrepo` (not bare `$HOME` default root).
- Shared **cold seed** run (no `-v`, `WRK_SCAN_DEBUG` forced empty) populates:
  - `{WRK_HOME}/projects.json` with the main (`source=scan`)
  - product mirror under `{FakeHome}/.cache/git-repo-scan` so the Run scan is warm-eligible
- Seed uses the same FakeHome/WRK_HOME as Run so cache and projects match.
- Leaves only differ on how debug is enabled for the **second** `wrk --scan-git-repos` (the Run under test).
- Cwd remains non-git `{WorkRoot}` (parent Setup).

## Steps

1. Set `req.FakeHome = req.WorkRoot`.
2. Create `{WorkRoot}/scan-root/myrepo` as a main git repo; remember `MainRepo`.
3. Cold-seed once: `wrk --scan-git-repos <scan-root>` with FakeHome env and `WRK_SCAN_DEBUG=` (no `-v`).
4. Descendants set Args / ExtraEnv for the debug-on or debug-off second run.

## Context

- Library contract (P2): `Options.Debug` emits phase-level `scan:` lines; warm/cold include `mode=…`.
- Wrk contract (P3): wire Debug from `-v` OR `WRK_SCAN_DEBUG`; pass `os.Stderr` as Stderr.
- Optional wrk-side log (not required by Assert): `scan: record known=N newly=M`.
- Ambient `WRK_SCAN_DEBUG` from the host must not leak into seed or `off/` (seed and off force empty).

```go
import (
	"os/exec"
	"path/filepath"
	"os"
)

// seedScanGitReposNoDebug runs a quiet first scan so projects.json and the
// product default mirror under FakeHome ($HOME/.cache/git-repo-scan) are
// populated. Does not enable -v or WRK_SCAN_DEBUG (forces env empty so ambient
// host env cannot pollute seed).
func seedScanGitReposNoDebug(t *testing.T, req *Request, scanRoot string) {
	t.Helper()
	if req.FakeHome == "" {
		t.Fatal("seedScanGitReposNoDebug requires FakeHome for product CacheRoot isolation")
	}
	bin := getWrkBin(t)
	cmd := exec.Command(bin, "--scan-git-repos", scanRoot)
	cmd.Dir = req.RepoDir
	env := wrkEnv(req)
	// Last WRK_SCAN_DEBUG= wins over ambient os.Environ() duplicates.
	env = append(env, "WRK_SCAN_DEBUG=")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("seed wrk --scan-git-repos: exit %d output=%q", ee.ExitCode(), string(out))
		}
		t.Fatalf("seed wrk --scan-git-repos: %v output=%q", err, string(out))
	}
	// Seed should have recorded the main at least once.
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")
}

// forceScanDebugEnvOff appends WRK_SCAN_DEBUG= so Run sees non-truthy env
// (covers ambient host pollution for off/ and via-verbose leaves).
func forceScanDebugEnvOff(req *Request) {
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SCAN_DEBUG=")
}

func Setup(t *testing.T, req *Request) error {
	// Isolate HOME → product default cache under sandbox only.
	req.FakeHome = req.WorkRoot

	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "myrepo")
	req.MainRepo = mainRepo

	// Quiet cold seed: populate projects + default cache (no debug flags).
	seedScanGitReposNoDebug(t, req, scanRoot)

	// Default leaf Args: second scan of the same explicit root (leaves override flags/env).
	req.Args = []string{"--scan-git-repos", scanRoot}

	// Keep helper referenced from this package root (leaves call it too).
	_ = forceScanDebugEnvOff
	_ = filepath.Join
	return nil
}
```
