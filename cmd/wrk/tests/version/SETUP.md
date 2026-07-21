# Scenario

**Feature**: wrk --version prints embedded build version

```
# no git repo required
wrk --version -> stdout v0.0.1\n (from go:embed VERSION.txt)

# help documents the flag
wrk -h -> usage mentions --version

# sole top-level flag only
wrk --version + other mode flag -> non-zero, mutually exclusive
```

## Preconditions

- The wrk Go module is at `go-pkgs/cmd/wrk/` (three levels above this tree root).
- Go toolchain is available on PATH.
- Session wrk binary is built once per doctest run to
  process-local `MkdirTemp` wrk binary (in-memory mutex).
- Embedded version lives in `wrkcli/VERSION.txt` (compiled into the binary).

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Descendants set `req.Args` for the wrk invocation from a neutral cwd.

## Context

- Version commands do not require a git repository; cwd is a neutral empty dir.
- Only `WRK_HOME` is passed via `versionWrkEnv`; no other wrk env overrides.
- Stdout assertions use `assert.Output` v2 full-match templates where output is
  bounded; help uses substring checks for `--version`.
- Printed version string is exactly `v0.0.1` (file content `0.0.1` plus `v` prefix).

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
	"sync"
)

func findModuleRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Process-local wrk binary (one-process suite; in-memory mutex, not session flock).
var (
	wrkBinMu   sync.Mutex
	wrkBinPath string
	wrkBinErr  error
	// wrkModRoot set from d.DOCTEST_ROOT in root Setup.
	wrkModRoot string
)

func getWrkBin(t *testing.T) string {
	t.Helper()
	wrkBinMu.Lock()
	defer wrkBinMu.Unlock()
	if wrkBinPath != "" || wrkBinErr != nil {
		if wrkBinErr != nil {
			t.Fatal(wrkBinErr)
		}
		return wrkBinPath
	}
	if wrkModRoot == "" {
		t.Fatal("wrkModRoot unset; root Setup must run first")
	}
	dir, err := os.MkdirTemp("", "wrk-doctest-bin-")
	if err != nil {
		wrkBinErr = err
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "wrk")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "./cmd/wrk")
	cmd.Dir = wrkModRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		wrkBinErr = fmt.Errorf("build wrk: %v\n%s", err, out)
		t.Fatal(wrkBinErr)
	}
	wrkBinPath = bin
	return bin
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if root := findModuleRoot(d.DOCTEST_ROOT); root != "" {
		wrkModRoot = root
	} else {
		t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
	}
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	if req.RepoDir == "" {
		req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	}
	ensureVersionHelpersUsed()
	return nil
}

func versionWrkEnv(req *Request) []string {
	return append(os.Environ(), "WRK_HOME="+req.WrkHome)
}

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", cwd, err)
	}
	return cwd
}

func eventsJSONLPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

func v2StdoutTemplate(body string) string {
	if body == "" {
		return "---\nversion: 2\n---\n"
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return "---\nversion: 2\n---\n" + body
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func assertVersionMutualExclusion(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "mutually exclusive") {
		return
	}
	if strings.Contains(lower, "unexpected") {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}

func ensureVersionHelpersUsed() {
	_ = eventsJSONLPath
	_ = assertVersionMutualExclusion
}
```