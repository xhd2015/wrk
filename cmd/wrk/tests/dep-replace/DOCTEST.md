# wrk --dep-replace — unwind-stack fan-out + CLI tree (absolute, no tidy)

## Version
0.0.2

Decision tree for **`wrk --dep-replace <dir>… [--dry-run]`**.

Adds (or rewrites) an **absolute** `replace module => absDir`. Consumer set =
**`CollectStackInventory(cwd)`** (cwd git toplevel + nested status repos +
BFS over local filesystem replaces). Scan **every** `go.mod` under every
member `Path`. Write only if the module already `require`s the dep path **or**
already has a `replace` for that path. Self (`consumer.Path == dep.Path`)
never rewritten. **Not git** → today’s walk-up nearest go.mod and **D7**
(write even with no require). Multi-dir: dry-run **validates every dir
first**; apply write errors are **fail-fast** (prior writes may remain).
**No tidy.** Never create go.sum.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
No L3 e2e leaves. Parallel-safe: inject Env/Dir via Capture.

**Locked product rules:**

| ID | Rule |
|----|------|
| D1 | Absolute replace NewPath (not relative; relative is `--pin-locals`) |
| D2 | **No tidy** (never create go.sum) |
| D3 | Apply multi-arg **fail-fast** on write errors (prior writes may remain). Dry-run: validate **every** dir first; any bad arg → no banner / no tree |
| D4 | No `--all` / `--replaced` |
| D5 | CLI tree (`====`, `dep`, `checkout`, `module`, `replace` / `would: replace`); no kool commit-message line |
| D6 | Git: `CollectStackInventory(cwd)`. Not git: nearest go.mod from workDir |
| D7 | Not-git nearest fallback still writes **without** a prior require. Git stack: gated (require **or** existing replace). Self never rewritten |
| D8 | Zero gated consumers on the whole stack (git) → `wrk:` error containing `replace` or `consumer`; **no banner** |

**Partners:** `--dry-run`, optional `--color` / `--no-color`.

**Out of scope:** `--dep-update` (P1 sealed); relative replace (`--pin-locals`);
tidy; `--all`; commit; kool.

# DSN (Domain Specific Notion)

- **wrk --dep-replace** — exclusive top-level mode. One or more dep directory
  args. Partners: `--dry-run`, color. Does not exec kool. No tidy.
- **Stack consumer set** — `CollectStackInventory(cwd)`: cwd git toplevel +
  nested status repos + BFS over local filesystem replaces (same as
  `--unwind` / `--pin-locals` / `--dep-update`). Scan every `go.mod` under
  every member `Path`. Not git → nearest `go.mod` (D7: write even with no
  require).
- **Gate (git)** — write absolute replace only if the module already requires
  the dep path **or** already has a replace for that path. Self never rewritten.
  Zero gated consumers on the stack → `wrk:` error; no banner.
- **Absolute replace** — `replace <module> => <absDir>` (never `./` / `../`).
- **CLI tree** — one banner; `dep` headers (argv order); body checkout →
  module → `replace` / `would: replace`. Checkout label = `statusDirLine`
  vs invocation cwd (`.` / `external/kool`).
- **Dry-run** — validate every dir arg first (any bad → no banner / no tree);
  then `would:` replace lines; zero writes.
- **Apply fail-fast** — write errors stop; earlier successful replaces may
  already be on disk.
- **Exclusive** — XOR with `--dep-update`; exclusive with other primary modes
  (`--pin-locals`, `--done`, `--unwind`, `--bring`, `--list`, …).
- **Hard errors** — empty paths; missing dir; dep not a go module; no consumer
  go.mod (not-git walk-up); zero gated stack consumers. No banner.

# Decision Tree

```text
dep-replace/
├── help/
│   ├── mentions-flag                  # --dep-replace
│   └── mentions-stack                 # unwind/stack for --dep-replace
├── reject/
│   ├── no-args
│   ├── with-dep-update
│   └── with-pin-locals
├── error/
│   ├── missing-dir
│   ├── not-a-module
│   ├── no-consumer-gomod              # not-git walk-up failed
│   └── no-stack-consumer              # git stack; zero gated consumers
├── dry-run/
│   ├── no-write
│   ├── stack-no-write
│   ├── multi-dir-stack-no-write
│   └── bad-second-arg                 # no banner; first dep not a half-plan
└── apply/
    ├── single-dir                     # not-git nearest
    ├── multi-dir
    ├── nested-module                  # workDir under consumer; walk up
    ├── no-existing-require            # D7 not-git nearest
    ├── fail-fast-second-missing       # apply: first write may stay
    ├── stack-other-checkout
    ├── stack-skip-self
    ├── stack-skip-non-consumer
    ├── stack-existing-replace         # other checkout gated by existing replace
    └── multi-dir-stack
```

# Test Index

| Path | Intent |
|------|--------|
| `help/mentions-flag` | Root help mentions `--dep-replace` |
| `help/mentions-stack` | Root help mentions unwind/stack (or equivalent) for `--dep-replace` |
| `reject/no-args` | Empty paths → error requires directory |
| `reject/with-dep-update` | XOR with `--dep-update` |
| `reject/with-pin-locals` | Exclusive with `--pin-locals` |
| `error/missing-dir` | Nonexistent dep path → non-zero |
| `error/not-a-module` | Dep dir without go.mod → non-zero |
| `error/no-consumer-gomod` | No go.mod above workDir (not-git) → non-zero; no banner |
| `error/no-stack-consumer` | Git stack with zero gated consumers → `replace`/`consumer` error; no banner |
| `dry-run/no-write` | Dry-run tree (`would: replace`); no go.mod write |
| `dry-run/stack-no-write` | Single-target stack tree; zero writes |
| `dry-run/multi-dir-stack-no-write` | Two `dep` headers; would: replace; zero writes |
| `dry-run/bad-second-arg` | No banner; `wrk:` + missing dir; first dep not a half-plan |
| `apply/single-dir` | Not-git nearest: tree + absolute replace; no tidy |
| `apply/multi-dir` | Two `dep` headers; both replaces; tidy-less; one consumer |
| `apply/nested-module` | workDir nested; nearest parent go.mod edited |
| `apply/no-existing-require` | Not-git D7: write without prior require |
| `apply/fail-fast-second-missing` | Apply: first replace stays; second missing → non-zero |
| `apply/stack-other-checkout` | Absolute replace on primary **and** other stack checkout |
| `apply/stack-skip-self` | Dep’s own go.mod not rewritten |
| `apply/stack-skip-non-consumer` | Other-checkout with neither require nor replace → untouched |
| `apply/stack-existing-replace` | Other checkout gated by existing replace (no require) |
| `apply/multi-dir-stack` | Two dep args; one consumer both replaces; other only first |

# Output contracts (assert targets)

Tokens: `====`, `dep`, `checkout`, `module`, `replace`, `would:`, `=>`.
Trailing `\n`. Checkout = `statusDirLine` vs invocation cwd (`.` /
`external/kool`). No tidy lines. No short form for a single target.

**Apply success (stdout, trailing `\n`):**

```text
==== dep-replace ====
dep  <dep-path> => <abs>

  checkout  .
    module  <consumer>
      replace  <dep-path> => <abs>

dep-replace: replaced in N modules in C checkouts
```

**Multiple targets:** N `dep` header lines (argv order). Each module lists
only replaces it is gated to receive.

**Dry-run:** banner `==== dep-replace (dry-run) ====`; `would: replace  …`;
summary `dep-replace: would replace in N modules in C checkouts`. Zero writes.

**Errors (stderr, non-zero):** `wrk:`; **no banner**. Bad second arg on
dry-run: `no such dir` (or equivalent); no tree; first dep not a half-plan.

**Exclusive / validation (stderr, non-zero):** prefer naming the flag;
empty-paths wording includes directory / requires; mutual exclusion wording
includes `mutually exclusive` (or equivalent).

No kool-style `commit message:` lines. Do not assert ANSI.

# How to Run

```sh
doctest vet ./cmd/wrk/tests/dep-replace
doctest test ./cmd/wrk/tests/dep-replace
```

Classic TDD: **new stack / new-stdout** leaves and **rewritten** success/dry-run
tree asserts stay **RED** until the implementer lands stack fan-out + CLI tree.
Reject/error leaves keep today’s meaning and may stay **GREEN**.

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

	// Extra consumer (stack other-checkout / skip-self).
	Consumer2ModDir     string
	Consumer2GoMod      string
	Baseline2GoMod      string
	WantConsumerModule  string
	WantConsumer2Module string
	WantUpdated         int // N in "replaced in N modules"
	WantCheckouts       int // C in "in C checkouts"; 0 → helper default 1
	WantCheckout        string
	WantCheckout2       string
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
