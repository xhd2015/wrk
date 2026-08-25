# Scenario

**Feature**: wrk bash tab-completion integration lifecycle and callback

```
# print script
wrk --bash-integration -> stdout bash completion script

# help (short-circuits; never dumps script / never mutates)
wrk --bash-integration -h|--help -> dedicated usage
wrk --bash-integration --install --help -> usage, no write

# install / uninstall / status (fake HOME + isolated WRK_HOME)
wrk --bash-integration --install -> bash.sh + dual profile markers
wrk --bash-integration --uninstall -> strip markers, keep bash.sh
wrk --bash-integration --status -> installed | not installed | partial

# completion callback
wrk --bash-integration --complete -- <words> <cword> -> basename/flag candidates
# path-like cur (/ ./ ../) -> empty candidates; script yields via -o default / compopt
```

## Preconditions

- The wrk Go module is two levels above this tree (`go-pkgs/cmd/wrk/`).
- Go toolchain is available on PATH.
- Session-built `wrk` binary at `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk`.
- Each leaf uses isolated `{WorkRoot}/.wrk` and fake `{WorkRoot}/home` for profile files.

## Steps

1. Root `Setup` ensures helper symbols are referenced (vet).
2. Descendants set `req.Mode`, pre-seed fixtures, and completion words.

## Context

- Profile targets: both `~/.bash_profile` and `~/.bashrc`.
- Script path: `{WRK_HOME}/integration/bash.sh`.
- Bash-integration modes skip `events.jsonl` append.
- User-facing stdout lines end with trailing `\n`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"

	"github.com/xhd2015/wrk/wrkcli/storage"
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
	ensureBashIntegrationHelpersUsed()
	return nil
}

func wrkMarkerBlock() string {
	return `# === wrk integration begin ===
_wrk_home="${WRK_HOME:-$HOME/.wrk}"
[[ -f "$_wrk_home/integration/bash.sh" ]] && source "$_wrk_home/integration/bash.sh"
# === wrk integration end ===
`
}

func minimalBashSh() string {
	return `#!/usr/bin/env bash
# wrk integration stub for doctest pre-seed
_wrk() { :; }
complete -o default -F _wrk wrk
`
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

func writeProjectsJSON(wrkHome string, paths []string) error {
	pf := storage.ProjectsFile{Version: 1}
	for _, p := range paths {
		pf.Projects = append(pf.Projects, storage.Project{
			Path:    p,
			AddedAt: time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Source:  storage.SourceManual,
		})
	}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(wrkHome, "projects.json"), data, 0o644)
}

func preInstalledProfileContent() string {
	return `# user config
export EDITOR=vim
` + wrkMarkerBlock()
}

func preInstalledBashShContent() string {
	return `#!/usr/bin/env bash
# pre-seeded wrk bash integration
_wrk() { :; }
complete -o default -F _wrk wrk
`
}

func assertHomeIsolated(t *testing.T, path, home string) {
	t.Helper()
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %q: %v", path, err)
	}
	homeAbs, err := filepath.Abs(home)
	if err != nil {
		t.Fatalf("abs home %q: %v", home, err)
	}
	sep := string(filepath.Separator)
	if !strings.HasPrefix(absPath, homeAbs+sep) && absPath != homeAbs {
		t.Fatalf("path %q is outside isolated HOME %q", absPath, homeAbs)
	}
}

func assertWrkHomeIsolated(t *testing.T, path, wrkHome string) {
	t.Helper()
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %q: %v", path, err)
	}
	wrkAbs, err := filepath.Abs(wrkHome)
	if err != nil {
		t.Fatalf("abs wrk home %q: %v", wrkHome, err)
	}
	sep := string(filepath.Separator)
	if !strings.HasPrefix(absPath, wrkAbs+sep) && absPath != wrkAbs {
		t.Fatalf("path %q is outside isolated WRK_HOME %q", absPath, wrkAbs)
	}
}

func assertStdoutEndsWithNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		return
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with trailing newline; got %q", stdout)
	}
}

func assertNoEventsJSONL(t *testing.T, resp *Response) {
	t.Helper()
	if _, err := os.Stat(resp.EventsPath); err == nil {
		t.Fatalf("bash-integration must not append events.jsonl at %s", resp.EventsPath)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q in %q", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("expected %q not in %q", substr, s)
	}
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
}

func ensureBashIntegrationHelpersUsed() {
	_ = wrkMarkerBlock
	_ = minimalBashSh
	_ = bashShPath
	_ = countWrkMarkers
	_ = readFileIfExists
	_ = writeProjectsJSON
	_ = preInstalledProfileContent
	_ = preInstalledBashShContent
	_ = assertHomeIsolated
	_ = assertWrkHomeIsolated
	_ = assertStdoutEndsWithNewline
	_ = assertNoEventsJSONL
	_ = assertContains
	_ = assertNotContains
	_ = requireMode
	_ = requireNoPreseed
}
```