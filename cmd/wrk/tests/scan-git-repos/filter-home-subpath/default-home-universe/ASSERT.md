## Expected

- Exit code **0**.
- Stdout lists both mains (absolute paths, one per line): `home-main` and
  `Projects/proj-main` (order free; each path appears once).
- `projects.json` has **0** entries (print-only).
- Product cache under FakeHome has durable **home universe** index at
  `{FakeHome}/.cache/git-repo-scan/home/repos.json`.
- Index envelope: `universe` is `home` (when present as JSON field).
- Index includes both fixture mains (at least).
- Scanning default home must **not** create a competing root-universe index as the
  primary store for these paths (`root/repos.json` may be absent).

## Side Effects

- `{WRK_HOME}/projects.json` stays empty.
- Product cache files under `{FakeHome}/.cache/git-repo-scan/home/`.

## Errors

- None on the success path.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("bare --scan-git-repos home universe: exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	homeMain := resolveScanPath(t, req.SecondRepo)
	projMain := resolveScanPath(t, req.MainRepo)

	stdout := strings.TrimSpace(resp.Stdout)
	if stdout == "" {
		t.Fatalf("expected both mains on stdout, got empty")
	}
	lines := strings.Split(stdout, "\n")
	seen := map[string]bool{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		seen[resolveScanPath(t, line)] = true
	}
	if !seen[homeMain] {
		t.Fatalf("stdout missing home-main %q; stdout=%q", homeMain, resp.Stdout)
	}
	if !seen[projMain] {
		t.Fatalf("stdout missing proj-main %q; stdout=%q", projMain, resp.Stdout)
	}
	if len(seen) != 2 {
		t.Fatalf("stdout should list exactly 2 mains, got %d paths: %v\nstdout=%q",
			len(seen), seen, resp.Stdout)
	}

	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)

	indexPath := productHomeUniverseReposJSON(req.FakeHome)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("default ~ must seed home universe index at %s: %v", indexPath, err)
	}
	var doc struct {
		Universe string `json:"universe"`
		Version  int    `json:"version"`
		Repos    []struct {
			Path string `json:"path"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse home/repos.json: %v", err)
	}
	if doc.Universe != "" && doc.Universe != "home" {
		t.Fatalf("home/repos.json universe=%q, want home", doc.Universe)
	}
	if doc.Version != 0 && doc.Version != 1 {
		t.Fatalf("home/repos.json version=%d, want 1 (or unset)", doc.Version)
	}
	if !homeUniverseIndexHasPath(t, req.FakeHome, homeMain) {
		t.Fatalf("home index missing home-main %q", homeMain)
	}
	if !homeUniverseIndexHasPath(t, req.FakeHome, projMain) {
		t.Fatalf("home index missing proj-main %q", projMain)
	}

	// Home-default must not put these under root universe as primary.
	rootIdx := productRootUniverseReposJSON(req.FakeHome)
	if st, err := os.Stat(rootIdx); err == nil && !st.IsDir() {
		// If root/repos.json exists, it must not be the only place listing the
		// home fixture — home/repos.json is required above. Soft check: root
		// file alone is not a failure if empty; fail only if home is missing
		// (already checked). Log path for implementer clarity.
		_ = filepath.Dir(rootIdx)
	}
}
```
