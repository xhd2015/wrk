# wrk --dep-update — drop replace + require latest git tag

## Version
0.0.2

Decision tree for **Classic TDD** greenfield mode **`wrk --dep-update <dir>…`**.

For each dep directory: drop the replace for that module in the **nearest
consumer go.mod** and set `require` to the **latest matching git tag** version.
Library: `gotool/update.Pin` with explicit `ConsumerDir` / `DepDir` (no
process Chdir). **No tidy.** Multi-arg **fail-fast**.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
No L3 e2e leaves. Leaves stay **RED** until implementer lands the flag.

**Locked product rules:**

| ID | Rule |
|----|------|
| D1 | Absolute replace is the replace-mode concern; update **drops** replace |
| D2 | **No tidy** |
| D3 | Multi-arg **fail-fast** |
| D4 | No `--all` / `--replaced` in v1 |
| D5 | Structured wrk lines only (`dep-update …`); no kool commit-message line |
| D6 | Nearest go.mod from workDir only |
| D7 | Replace mode does not require prior require; update sets require to tag |

**Partners:** `--dry-run`, optional `--color` / `--no-color`.

**Out of scope v1:** multi-module fan-out, forced Version flag, tidy, kool shell-out.

# DSN (Domain Specific Notion)

- **wrk --dep-update** — exclusive top-level mode. Takes one or more dep
  directory args. For each: resolve module + latest tag under DepDir, drop
  replace in nearest consumer go.mod, set require to `vN.N.N`.
- **Nearest consumer** — walk up from workDir to nearest go.mod; edit that only.
- **Tag resolution** — latest numeric tag with submodule prefix when dep is
  nested under a monorepo (`packages/dep/v0.0.2` → version `v0.0.2`).
- **Dry-run** — print `would: dep-update …` only; zero go.mod writes.
- **Apply** — drop replace + edit require; print
  `dep-update <module> -> <version>` (optional tag parenthetical
  implementer-owned); **no** tidy.
- **Exclusive** — XOR with `--dep-replace`; exclusive with other primary modes.
- **Hard errors** — empty paths; missing dir; not a go module; **no tags**;
  no consumer go.mod.

# Decision Tree

```text
dep-update/
├── help/mentions-flag
├── reject/
│   ├── no-args
│   ├── with-dep-replace
│   └── with-pin-locals
├── error/
│   ├── missing-dir
│   ├── not-a-module
│   ├── no-tags
│   └── no-consumer-gomod
├── dry-run/no-write
└── apply/
    ├── drop-replace-set-require      # has local replace + old require
    ├── nested-module-tag-prefix      # packages/dep/vN.N.N tags
    └── multi-dir
```

# Test Index

| Path | Intent |
|------|--------|
| `help/mentions-flag` | Root help mentions `--dep-update` |
| `reject/no-args` | Empty paths → requires directory |
| `reject/with-dep-replace` | XOR with `--dep-replace` |
| `reject/with-pin-locals` | Exclusive with `--pin-locals` |
| `error/missing-dir` | Nonexistent dep → non-zero |
| `error/not-a-module` | Dir without go.mod → non-zero |
| `error/no-tags` | Module with no version tags → non-zero |
| `error/no-consumer-gomod` | No go.mod up walk → non-zero |
| `dry-run/no-write` | `would: dep-update …`; no go.mod write |
| `apply/drop-replace-set-require` | Drop replace; require@latest tag |
| `apply/nested-module-tag-prefix` | Submodule tag prefix → clean version |
| `apply/multi-dir` | Two deps updated fail-fast success |

# Output contracts (assert targets)

**Apply success (stdout, trailing `\n`):**

```text
dep-update <module-path> -> v0.0.2
```

Optional tag parenthetical is implementer-owned, e.g.
`(tag packages/dep/v0.0.2)`. Locked tokens: `dep-update`, `->`, version
`vN.N.N`.

**Dry-run:**

```text
would: dep-update <module-path> -> v0.0.2
```

**Exclusive / validation (stderr, non-zero):** mutual exclusion wording;
no-tags / missing / not-module wording greppable.

# How to Run

```sh
doctest vet ./cmd/wrk/tests/dep-update
doctest test ./cmd/wrk/tests/dep-update
```

Classic TDD: expect **RED** until `--dep-update` is implemented.

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

// Request drives wrk --dep-update under isolated WorkRoot / WRK_HOME.
type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string
	Args     []string
	ExtraEnv []string

	InProcess bool

	ConsumerModDir string
	ConsumerGoMod  string
	DepDir         string
	Dep2Dir        string
	MissingPath    string
	BaselineGoMod  string

	// Expected pin outcomes (seeded by fixtures).
	WantVersion  string // e.g. v0.0.2
	WantVersion2 string
	WantTagHint  string // optional substring of tag form e.g. packages/dep/v0.0.2
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	adoptDoctestContext(d)
	args := append([]string(nil), req.Args...)
	dir := req.RepoDir
	if dir == "" {
		dir = req.WorkRoot
	}

	if req.InProcess {
		res := wrkcli.Capture(wrkcli.CaptureOpts{
			Args: args,
			Dir:  dir,
			Env:  depUpdateEnv(req),
		})
		return &Response{
			Stdout:   res.Stdout,
			Stderr:   res.Stderr,
			ExitCode: res.ExitCode,
		}, nil
	}

	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = depUpdateEnv(req)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

func depUpdateEnv(req *Request) []string {
	base := filterEnvKeys(os.Environ(), "NO_COLOR")
	env := append(base,
		"WRK_HOME="+req.WrkHome,
		"WRK_DATE="+wrkDate,
	)
	if len(req.ExtraEnv) > 0 {
		env = append(env, req.ExtraEnv...)
	}
	return env
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
		modRoot := findModuleRoot(doctestRootPath(t))
		if modRoot == "" {
			t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
		}
		if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
			t.Fatalf("mkdir bin dir: %v", err)
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

var (
	harnessMu          sync.Mutex
	harnessSessionID   string
	harnessDoctestRoot string
)

func adoptDoctestContext(d *session.Doctest) {
	if d == nil {
		return
	}
	harnessMu.Lock()
	defer harnessMu.Unlock()
	if d.DOCTEST_SESSION_ID != "" {
		harnessSessionID = d.DOCTEST_SESSION_ID
	}
	if d.DOCTEST_ROOT != "" {
		harnessDoctestRoot = d.DOCTEST_ROOT
	}
}

func doctestSessionID(t *testing.T) string {
	t.Helper()
	harnessMu.Lock()
	sid := harnessSessionID
	harnessMu.Unlock()
	if sid == "" {
		t.Fatal("d.DOCTEST_SESSION_ID not set (expected adoptDoctestContext from Setup)")
	}
	return sid
}

func doctestRootPath(t *testing.T) string {
	t.Helper()
	harnessMu.Lock()
	root := harnessDoctestRoot
	harnessMu.Unlock()
	if root == "" {
		t.Fatal("d.DOCTEST_ROOT not set (expected adoptDoctestContext from Setup)")
	}
	return root
}

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
	return filepath.Join(fixtureCacheBase(t), doctestSessionID(t))
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

func filterEnvKeys(env []string, drop ...string) []string {
	dropSet := make(map[string]struct{}, len(drop))
	for _, k := range drop {
		dropSet[k] = struct{}{}
	}
	out := make([]string, 0, len(env))
	for _, e := range env {
		key := e
		if i := strings.IndexByte(e, '='); i >= 0 {
			key = e[:i]
		}
		if _, skip := dropSet[key]; skip {
			continue
		}
		out = append(out, e)
	}
	return out
}

var _ = fmt.Sprintf
```
