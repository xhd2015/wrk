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
  `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk`.
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
	"syscall"
	"testing"

	"github.com/xhd2015/doctest/assert"
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

func fixtureCacheBase(t *testing.T) string {
	t.Helper()
	base := os.Getenv("DOCTEST_FIXTURE_ROOT")
	if base != "" {
		return base
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(home, "Library", "Caches", "doctest", "fixtures")
}

func fixtureSessionRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join(fixtureCacheBase(t), DOCTEST_SESSION_ID)
}

func sessionWrkBin(t *testing.T) string {
	t.Helper()
	return filepath.Join(fixtureSessionRoot(t), "bin", "wrk")
}

func withFlock(t *testing.T, lockPath string, fn func()) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("mkdir lock dir: %v", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("open lock %s: %v", lockPath, err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock %s: %v", lockPath, err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()
	fn()
}

func getWrkBin(t *testing.T) string {
	t.Helper()
	bin := sessionWrkBin(t)
	if _, err := os.Stat(bin); err == nil {
		return bin
	}
	lockPath := filepath.Join(fixtureSessionRoot(t), "bin", ".lock")
	withFlock(t, lockPath, func() {
		if _, err := os.Stat(bin); err == nil {
			return
		}
		modRoot := findModuleRoot(DOCTEST_ROOT)
		if modRoot == "" {
			t.Fatal("find module root: no go.mod in ancestors")
		}
		cmd := exec.Command("go", "build", "-o", bin, "./cmd/wrk")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build wrk: %v\n%s", err, out)
		}
	})
	return bin
}

func Setup(t *testing.T, req *Request) error {
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