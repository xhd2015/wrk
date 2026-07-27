# Scenario

**Feature**: wrk --bash-integration --install reports per-component status

```
# real install writes script + dual profile markers, then prints report
fake HOME + WRK_HOME
wrk --bash-integration --install
  -> integration/bash.sh + markers
  -> stdout: bash integration: <summary> + component lines

# dry-run previews with would-* vocabulary, no writes
wrk --bash-integration --install --dry-run
  -> stdout only (would install | would update | is up to date)
```

## Preconditions

- The wrk Go module main package is two levels above this tree (`go-pkgs/cmd/wrk/`).
- Go toolchain is available on PATH.
- Session-built `wrk` binary at `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk`
  (file-locked across leaf processes).
- Each leaf uses isolated `{WorkRoot}/.wrk` and fake `{WorkRoot}/home`.

## Steps

1. Root `Setup` allocates WorkRoot / WRK_HOME / FakeHome and registers helpers.
2. Descendants set `req.Mode`, dry-run, and pre-seed fixtures.

## Context

- Report shape (real install): four lines + trailing blank line.
- Script status: `installed` | `updated` | `is up to date`.
- Marker status: `installed` | `is up to date` (never `updated`).
- Summary: `installed` if script was missing; else `updated` if script updated
  or any marker installed; else `is up to date`.
- Dry-run uses `would install` / `would update` / `is up to date` with the same
  four-line layout (no mandatory `dry-run:` prefix). Marker-block preview after
  status lines is out of scope for this tree (not asserted; prefer absent).
- Bash-integration install skips `events.jsonl`.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

const (
	wrkMarkerBegin = "# === wrk integration begin ==="
	wrkMarkerEnd   = "# === wrk integration end ==="
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	if req.WrkHome == "" {
		req.WrkHome = filepath.Join(workRoot, ".wrk")
	}
	if req.FakeHome == "" {
		req.FakeHome = filepath.Join(workRoot, "home")
	}
	if req.RepoDir == "" {
		req.RepoDir = workRoot
	}
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.FakeHome, 0o755); err != nil {
		return err
	}
	ensureMessagingHelpersUsed()
	return nil
}

func wrkMarkerBlock() string {
	return `# === wrk integration begin ===
_wrk_home="${WRK_HOME:-$HOME/.wrk}"
[[ -f "$_wrk_home/integration/bash.sh" ]] && source "$_wrk_home/integration/bash.sh"
# === wrk integration end ===
`
}

func outdatedBashShContent() string {
	return `#!/usr/bin/env bash
# pre-seeded outdated wrk bash integration (must be rewritten)
_wrk() { :; }
complete -o default -F _wrk wrk
`
}

func preInstalledProfileContent() string {
	return `# user config
export EDITOR=vim
` + wrkMarkerBlock()
}

func bashShPath(wrkHome string) string {
	return filepath.Join(wrkHome, "integration", "bash.sh")
}

func countWrkMarkers(content string) int {
	return strings.Count(content, wrkMarkerBegin)
}

func readFileIfExists(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func requireMode(t *testing.T, req *Request, mode string) {
	t.Helper()
	if req.Mode != mode {
		t.Fatalf("expected mode %q, got %q", mode, req.Mode)
	}
}

func requireNoPreseed(t *testing.T, req *Request) {
	t.Helper()
	if req.PreExistingBashSh != "" || req.PreExistingBashProfile != "" || req.PreExistingBashRC != "" {
		t.Fatalf("expected no pre-seeded integration state")
	}
	if req.SeedCurrentScript || req.PreInstall {
		t.Fatalf("expected no pre-install / seed-current for fresh scenario")
	}
}

func assertNoEventsJSONL(t *testing.T, resp *Response) {
	t.Helper()
	if _, err := os.Stat(resp.EventsPath); err == nil {
		t.Fatalf("bash-integration must not append events.jsonl at %s", resp.EventsPath)
	}
}

func assertExit0(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s\nstdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertInstallReport(t *testing.T, resp *Response, summary, scriptStatus, profileMarkerStatus, rcMarkerStatus string) {
	t.Helper()
	tmpl := fmt.Sprintf(`---
version: 2
---
bash integration: %s
script: %s (%s)
bash_profile: %s (marker %s)
bashrc: %s (marker %s)

`, summary, resp.BashShPath, scriptStatus, resp.BashProfilePath, profileMarkerStatus, resp.BashRCPath, rcMarkerStatus)
	assert.Output(t, resp.Stdout, tmpl)
}

func assertDryRunUnchanged(t *testing.T, resp *Response) {
	t.Helper()
	if resp.BeforeBashShExists {
		if resp.BashShContent != resp.BeforeBashShContent {
			t.Fatalf("dry-run must not change bash.sh")
		}
	} else if _, err := os.Stat(resp.BashShPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create bash.sh at %s", resp.BashShPath)
	}
	if resp.BeforeBashProfileExists {
		if resp.BashProfileContent != resp.BeforeBashProfileContent {
			t.Fatalf("dry-run must not change .bash_profile")
		}
	} else if _, err := os.Stat(resp.BashProfilePath); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create .bash_profile at %s", resp.BashProfilePath)
	}
	if resp.BeforeBashRCExists {
		if resp.BashRCContent != resp.BeforeBashRCContent {
			t.Fatalf("dry-run must not change .bashrc")
		}
	} else if _, err := os.Stat(resp.BashRCPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create .bashrc at %s", resp.BashRCPath)
	}
}

func assertMarkersInstalled(t *testing.T, resp *Response) {
	t.Helper()
	if resp.BashProfileMarkerCount != 1 {
		t.Fatalf("expected 1 marker in .bash_profile, got %d:\n%s", resp.BashProfileMarkerCount, resp.BashProfileContent)
	}
	if resp.BashRCMarkerCount != 1 {
		t.Fatalf("expected 1 marker in .bashrc, got %d:\n%s", resp.BashRCMarkerCount, resp.BashRCContent)
	}
}

func ensureMessagingHelpersUsed() {
	_ = wrkMarkerBlock
	_ = outdatedBashShContent
	_ = preInstalledProfileContent
	_ = bashShPath
	_ = countWrkMarkers
	_ = readFileIfExists
	_ = requireMode
	_ = requireNoPreseed
	_ = assertNoEventsJSONL
	_ = assertExit0
	_ = assertInstallReport
	_ = assertDryRunUnchanged
	_ = assertMarkersInstalled
	_ = wrkMarkerEnd
}
```
