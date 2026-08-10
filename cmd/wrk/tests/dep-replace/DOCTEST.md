# wrk --dep-replace — absolute replace into nearest consumer go.mod

## Version
0.0.2

Decision tree for **Classic TDD** greenfield mode **`wrk --dep-replace <dir>…`**.

Adds (or rewrites) an **absolute** `replace module => absDir` in the **nearest
consumer go.mod** found by walking up from wrk `workDir`. Multi-dir args are
fail-fast. **No tidy.** Does **not** require an existing `require` for the
module (D7). Library surface: `gotool/replace.ReplaceIn` with explicit
`ConsumerDir` (no process Chdir).

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
No L3 e2e leaves. No product wiring yet — leaves stay **RED**.

**Locked product rules:**

| ID | Rule |
|----|------|
| D1 | Absolute replace NewPath (not relative; relative is `--pin-locals`) |
| D2 | **No tidy** |
| D3 | Multi-arg **fail-fast** (stop on first error; prior writes may remain) |
| D4 | No `--all` / `--replaced` in v1 |
| D5 | Structured wrk lines only (`dep-replace …`); no kool commit-message line |
| D6 | Nearest go.mod from workDir only |
| D7 | Replace does **not** require existing require |

**Partners:** `--dry-run`, optional `--color` / `--no-color`.

**Out of scope v1:** multi-module fan-out, relative replace, tidy, shelling out
to kool, `--all` / `--replaced`.

# DSN (Domain Specific Notion)

- **wrk --dep-replace** — exclusive top-level mode. Takes one or more dep
  directory args (`StringSlice`, repeated flags OK). For each arg: resolve
  module path from dep's go.mod, write absolute replace into the nearest
  consumer go.mod above workDir.
- **Nearest consumer** — walk up from process workDir until a `go.mod` is found;
  edit that module only.
- **Absolute replace** — `replace <module> => <absDir>` (never `./` / `../`).
- **Dry-run** — print `would: dep-replace …` only; zero go.mod writes.
- **Apply** — mutate consumer go.mod; print `dep-replace <module> => <abs>`;
  **no** `go mod tidy`.
- **Fail-fast multi-arg** — process dirs in order; first error stops; earlier
  successful replaces may already be on disk.
- **Exclusive** — XOR with `--dep-update`; exclusive with other primary modes
  (`--pin-locals`, `--done`, `--unwind`, `--bring`, `--list`, …).
- **Hard errors** — empty paths; missing dir; dep not a go module; no consumer
  go.mod up the walk.

# Decision Tree

```text
dep-replace/
├── help/mentions-flag                 # wrk -h documents --dep-replace
├── reject/                            # CLI / mode conflicts (minimal fixture)
│   ├── no-args                        # --dep-replace with zero dirs
│   ├── with-dep-update                # XOR partner
│   └── with-pin-locals                # exclusive family
├── error/                             # semantic preflight with fixtures
│   ├── missing-dir
│   ├── not-a-module
│   └── no-consumer-gomod
├── dry-run/no-write                   # would: line; go.mod unchanged
└── apply/                             # mutate consumer go.mod
    ├── single-dir
    ├── multi-dir
    ├── nested-module                  # workDir under consumer; walk up
    ├── no-existing-require            # D7: no prior require
    └── fail-fast-second-missing       # first applied, second fails
```

# Test Index

| Path | Intent |
|------|--------|
| `help/mentions-flag` | Root help mentions `--dep-replace` |
| `reject/no-args` | Empty paths → error requires directory |
| `reject/with-dep-update` | XOR with `--dep-update` |
| `reject/with-pin-locals` | Exclusive with `--pin-locals` |
| `error/missing-dir` | Nonexistent dep path → non-zero |
| `error/not-a-module` | Dep dir without go.mod → non-zero |
| `error/no-consumer-gomod` | No go.mod above workDir → non-zero |
| `dry-run/no-write` | `would: dep-replace …`; no go.mod write |
| `apply/single-dir` | Absolute replace written; stdout line |
| `apply/multi-dir` | Two replaces, two lines, exit 0 |
| `apply/nested-module` | workDir nested; nearest parent go.mod edited |
| `apply/no-existing-require` | Replace without prior require (D7) |
| `apply/fail-fast-second-missing` | First replace stays; second missing → non-zero |

# Output contracts (assert targets)

**Apply success (stdout, trailing `\n`):**

```text
dep-replace <module-path> => <absolute-path>
```

**Dry-run:**

```text
would: dep-replace <module-path> => <absolute-path>
```

Exact spacing implementer-owned; tokens `dep-replace`, `would:`, `=>` locked.
No kool-style `commit message:` lines.

**Exclusive / validation (stderr, non-zero):** prefer naming the flag;
empty-paths wording includes directory / requires; mutual exclusion wording
includes `mutually exclusive` (or equivalent).

# How to Run

```sh
doctest vet ./cmd/wrk/tests/dep-replace
doctest test ./cmd/wrk/tests/dep-replace
```

Classic TDD: expect **RED** until `--dep-replace` is implemented.

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

// Request drives wrk --dep-replace under isolated WorkRoot / WRK_HOME.
// Root Setup allocates WorkRoot / WrkHome. Leaves seed go.mod fixtures,
// set RepoDir (virtual cwd), Args, and path fields for asserts.
type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // virtual process cwd for Capture (consumer or nested)
	Args     []string
	ExtraEnv []string

	// InProcess runs via wrkcli.Capture (L2). Prefer true for all leaves.
	InProcess bool

	// Fixture paths / baselines for side-effect asserts.
	ConsumerModDir string
	ConsumerGoMod  string
	DepDir         string // primary dep module dir (absolute)
	Dep2Dir        string // second dep for multi-dir
	MissingPath    string // nonexistent path for error / fail-fast
	BaselineGoMod  string // go.mod snapshot before Run
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
			Env:  depReplaceEnv(req),
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
	cmd.Env = depReplaceEnv(req)

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

// depReplaceEnv isolates WRK_HOME / WRK_DATE; strips ambient NO_COLOR.
func depReplaceEnv(req *Request) []string {
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
