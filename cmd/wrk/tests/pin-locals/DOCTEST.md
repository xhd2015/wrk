# wrk --pin-locals — reverse peel: relative replace for stack deps already required

## Version
0.0.2

Decision tree for **Classic TDD** greenfield mode **`wrk --pin-locals`**.

Scan the **unwind stack inventory only** (primary + nested status repos + local
filesystem replace BFS — `CollectStackInventory` semantics; **not** WRK_HOME
project universe). Among stack modules, for each consumer, add or normalize
**relative** `replace` directives for dependencies that are **already** declared
(`require` including `// indirect`, or existing `replace` OldPath) and whose
module path is owned by some inventory module.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
No L3 e2e leaves. Pure helpers (`PlanPinLocals`, relative path) may land later;
this tree locks the **CLI surface**.

**Locked product rules (must stay RED until implementer lands):**

| ID | Rule |
|----|------|
| D1 | Inventory = unwind stack only |
| D2 | Conflicting / absolute replace → rewrite to relative stack owner |
| D3 | After consumer edits: `go mod tidy`; tidy fail → `warning:` + continue; summary stats; exit **0** on soft tidy fails |
| D4 | Multi-owner: same checkout / under consumer / stable first + warning |
| D5 | Nothing to do → already-up-to-date style + 0 applied, exit 0 |
| D6 | **Exclusive** mode — reject other modes **and** `--commit` / `--add-all` |
| D7 | Flag `--pin-locals`; partners `--dry-run`, `--color` / `--no-color` |

**Out of scope:** create worktrees; compose with commit/add-all/unwind/done;
WRK_HOME global pin universe; JSON output; rewriting sealed tests in other trees.

# DSN (Domain Specific Notion)

- **wrk --pin-locals** — exclusive top-level mode. Scans unwind **stack** from
  cwd, plans relative replaces for already-dependencies whose owners live in
  the stack, applies (or dry-runs) without creating worktrees or committing.
- **Stack inventory** — primary git toplevel + nested independent repos (status
  scan, typically `external/*`) + BFS over local filesystem replaces on modules
  under known checkouts. Intra-repo nested modules stay in the same checkout.
- **Wanted set (per consumer module)** — `require` paths (incl. indirect) ∪
  existing `replace` OldPaths. Never pin every co-located inventory module.
- **Relative replace** — all written `NewPath` values are `./` or `../` slash
  form (never absolute). Absolute same-target rewrites to relative.
- **Dry-run** — print `would: pin-local …` only; no go.mod writes; no tidy.
- **Apply** — edit replaces → `go mod tidy` per consumer needing work; continue
  on tidy failure with `warning:`; print success lines + end summary stats.
- **Exclusive** — mutually exclusive with other modes (`--done`, `--unwind`,
  `--bring`, `--list`, …) and with **`--commit`** / **`--add-all`**.
- **Hard errors** — not a git repository; exclusive/preflight. Soft tidy fails
  never force non-zero alone.

# Decision Tree

```text
pin-locals/
├── help/mentions-flag              # wrk -h documents --pin-locals
├── reject/                         # exclusive mode conflicts
│   ├── with-done
│   ├── with-unwind
│   ├── with-bring
│   ├── with-list
│   ├── with-commit                 # partners rejected even without other modes
│   └── with-add-all
├── error/not-git                   # non-git cwd → Error, non-zero
├── dry-run/                        # --pin-locals --dry-run
│   ├── multi-repo-external         # would: pin-local … => ../external/… or ./external/…
│   ├── already-up-to-date          # correct relative already present
│   └── skip-not-a-dependency       # inventory-only module → no would line
└── apply/                          # --pin-locals (mutate)
    ├── multi-repo-external         # relative replace written; not absolute
    ├── intra-project-nested        # root → ./tools
    ├── rewrite-absolute            # abs replace → relative same target
    ├── already-relative            # already correct → 0 applied / already message
    ├── tidy-fail-continues         # warning: + other modules still pinned; exit 0
    ├── idempotent-second-run       # apply twice; second already
    ├── skip-not-a-dependency       # no replace for non-required inventory module
    └── skip-no-matching-owner      # require path not in stack → no pin
```

# Test Index

| Path | Intent |
|------|--------|
| `help/mentions-flag` | Root help mentions `--pin-locals` |
| `reject/with-done` | Exclusive with `--done` |
| `reject/with-unwind` | Exclusive with `--unwind` |
| `reject/with-bring` | Exclusive with `--bring` |
| `reject/with-list` | Exclusive with `--list` |
| `reject/with-commit` | Rejects `--commit` |
| `reject/with-add-all` | Rejects `--add-all` |
| `error/not-git` | Non-git cwd hard fail |
| `dry-run/multi-repo-external` | would: pin + no go.mod write |
| `dry-run/already-up-to-date` | already message, 0 would-apply |
| `dry-run/skip-not-a-dependency` | no would for non-dep inventory |
| `apply/multi-repo-external` | relative replace + tidy soft path |
| `apply/intra-project-nested` | `./tools` replace |
| `apply/rewrite-absolute` | abs → relative rewrite |
| `apply/already-relative` | already up to date on apply |
| `apply/tidy-fail-continues` | soft tidy fail continue + stats exit 0 |
| `apply/idempotent-second-run` | second apply no-op |
| `apply/skip-not-a-dependency` | inventory-only skipped |
| `apply/skip-no-matching-owner` | off-stack require skipped |

# Output contracts (assert targets)

**Apply success (stdout):**

```text
pin-local <consumer-module-path> <- <dep-module-path> => <relative-path>
```

**Dry-run:**

```text
would: pin-local <consumer-module-path> <- <dep-module-path> => <relative-path>
```

**Tidy soft-fail (stderr):**

```text
warning: go mod tidy in <mod-dir>: …
```

**Summary (stdout, end of successful apply)** — must include explicit numbers
distinguishing success vs tidy fail. Locked preferred form:

```text
pin-locals: applied N, tidy ok M, tidy failed F
```

**Already up to date:** short message mentioning already / up to date; applied 0;
exit 0.

**Exclusive (stderr, non-zero):** `mutually exclusive` (or equivalent); prefer
naming `--pin-locals`.

**Not git (stderr, non-zero):** `not a git repository`.

# How to Run

```sh
doctest vet ./cmd/wrk/tests/pin-locals
doctest test ./cmd/wrk/tests/pin-locals
```

Classic TDD: expect **RED** until `--pin-locals` is implemented.

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

// Request drives wrk --pin-locals under isolated WorkRoot / WRK_HOME.
// Root Setup allocates WorkRoot / WrkHome. Leaves seed git+Go fixtures,
// set RepoDir (process cwd), Args, and fixture path fields for asserts.
type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // process cwd for wrk (git primary or plain non-git)
	Args     []string
	ExtraEnv []string

	// InProcess runs via wrkcli.Capture (L2). Prefer true for all leaves.
	InProcess bool

	// Fixture paths / baselines for side-effect asserts.
	ConsumerModDir string // consumer module dir (go.mod parent)
	ConsumerGoMod  string // path to consumer go.mod
	DepModDir      string // stack dep module dir
	ToolsModDir    string // intra-project nested module dir
	BadModDir      string // tidy-fail consumer module dir
	OtherModDir    string // inventory-only non-dep module dir
	BaselineGoMod  string // go.mod content snapshot before Run (dry-run / rewrite)
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
			Env:  pinLocalsEnv(req),
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
	cmd.Env = pinLocalsEnv(req)

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

// pinLocalsEnv isolates WRK_HOME / WRK_DATE; strips ambient NO_COLOR so leaves
// own color policy (pipe = plain). ExtraEnv appended last.
func pinLocalsEnv(req *Request) []string {
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

// getWrkBin builds session-cached wrk once (dual-mode binary path; unused when
// all leaves set InProcess=true).
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

// Silence unused when dual-mode helpers are only referenced from Run paths.
var _ = fmt.Sprintf
```
