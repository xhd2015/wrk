# wrk --dep-update — dir pin + inventory pull (`--all`)

## Version
0.0.2

Decision tree for **`wrk --dep-update`** in two modes:

1. **Dir mode** (existing, sealed): `wrk --dep-update <dir>… [--dry-run]` —
   for each dep directory, drop replace + set require to latest matching git tag
   via `gotool/update.Pin`. **No tidy.** Multi-arg **fail-fast.**
2. **`--all` inventory pull** (Classic TDD greenfield): `wrk --dep-update --all
   [--dry-run]` — scan all go.mod under **git toplevel of cwd**, match requires
   against `BuildInventory(WRK_HOME)`, pin inventory-owned deps to latest tags,
   then **`go mod tidy` once per affected consumer module**.

**CLI migration:** `--dep-update` is a **Bool** flag; dep directories are remaining
args (dir mode UX unchanged: `wrk --dep-update <dir>…`). Partner **`--all`**
(Bool). Rejects: neither dirs nor `--all`; `--all` with path args; bare `--all`
without `--dep-update`; exclusive with `--pin-locals` / `--dep-replace` / other
primary modes.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
No L3 e2e leaves. Parallel-safe: inject Env/Dir via Capture; no process env/cwd
mutation in harness.

**Locked product rules (dir mode):**

| ID | Rule |
|----|------|
| D1 | Absolute replace is the replace-mode concern; update **drops** replace |
| D2 | **No tidy** (dir mode only) |
| D3 | Multi-arg **fail-fast** |
| D4 | Dir mode does not use inventory / `--all` |
| D5 | Structured wrk lines only (`dep-update …`); no kool commit-message line |
| D6 | Nearest go.mod from workDir only |
| D7 | Replace mode does not require prior require; update sets require to tag |

**Locked product rules (`--all`):**

| ID | Rule |
|----|------|
| A1 | Consumer root = **git toplevel of cwd** (`worktree.ShowToplevel`), not main repo |
| A2 | Scan all go.mod under that toplevel; mutate **only** those modules |
| A3 | Inventory read-only (`BuildInventory`); ownership + latest tags from owner main path |
| A4 | External require → **silent skip** (does **not** increment `skipped S`); same-toplevel + filesystem replace → **skipped** (counts in S; no bump) |
| A5 | No tag on owner → `warning:` stderr + skip (counts in S); exit 0 |
| A6 | Already at latest → **already** count (no per-dep noise) |

| A7 | After apply pins: **`go mod tidy` once per consumer module with ≥1 pin**; no commit/build |
| A8 | Dry-run: `would:` plan + tidy lines; zero writes |
| A9 | Consumer need **not** be registered (pull-only) |

**Partners:** `--dry-run`, optional `--color` / `--no-color`, `-v`.

**Out of scope:** `--dep-replace --all`; commit/build gate; dir-mode tidy; JSON;
editing other projects' go.mod.

# DSN (Domain Specific Notion)

- **wrk --dep-update** — exclusive top-level Bool mode. Dir args (remaining) **or**
  partner **`--all`**. Partners: `--dry-run`, color, `-v`.
- **Dir mode** — for each dep dir: resolve module + latest tag under DepDir, drop
  replace in nearest consumer go.mod, set require to `vN.N.N`. No tidy.
- **`--all` mode** — resolve git toplevel of cwd; scan consumer modules under it;
  for each require, consult inventory ownership and owner tags; Pin when outdated
  inventory-owned; external require silent (not in `skipped S`); same-toplevel
  filesystem replace + no-tag soft-skips count in `skipped S`; tidy affected
  consumers after apply.
- **Inventory** — `BuildInventory(WRK_HOME)`: registered projects + sub-modules;
  latest numeric tags from owner on-disk module dir (main path).
- **Nearest consumer (dir)** — walk up from workDir to nearest go.mod.
- **Dry-run** — plan only; `would:` lines; zero go.mod writes.
- **Hard errors** — empty mode (no dirs, no `--all`); `--all` with paths; bare
  `--all`; exclusive modes; dir-mode missing dir / not module / no tags / no
  consumer; Pin/tidy write failures (fail-fast).
- **Soft warnings (`--all`)** — missing registry paths, no-tag skips (`warning:`).

# Decision Tree

```text
dep-update/
├── help/
│   ├── mentions-flag                # --dep-update (GREEN)
│   └── mentions-all                 # --all with --dep-update (RED until help)
├── reject/
│   ├── no-args                      # neither dirs nor --all
│   ├── with-dep-replace
│   ├── with-pin-locals
│   └── all/
│       ├── bare-all                 # --all without --dep-update
│       └── all-with-dirs            # --dep-update --all + path
├── error/
│   ├── missing-dir
│   ├── not-a-module
│   ├── no-tags
│   └── no-consumer-gomod
├── dry-run/no-write                 # dir mode
├── apply/                           # dir mode
│   ├── drop-replace-set-require
│   ├── nested-module-tag-prefix
│   └── multi-dir
└── all/
    ├── dry-run/
    │   ├── bumps-outdated
    │   └── already-up-to-date
    ├── apply/
    │   ├── cross-project-bump-and-tidy
    │   ├── worktree-toplevel-not-main
    │   ├── skip-intra-local-replace
    │   ├── skip-external-require
    │   └── nested-owner-module-tag
    └── soft/
        └── no-tag-warn
```

# Test Index

| Path | Intent |
|------|--------|
| `help/mentions-flag` | Root help mentions `--dep-update` |
| `help/mentions-all` | Root help mentions `--all` in context of `--dep-update` |
| `reject/no-args` | Neither dirs nor `--all` → requires directory or `--all` |
| `reject/with-dep-replace` | XOR with `--dep-replace` |
| `reject/with-pin-locals` | Exclusive with `--pin-locals` |
| `reject/all/bare-all` | Bare `--all` without `--dep-update` → error |
| `reject/all/all-with-dirs` | `--dep-update --all` with path args → error |
| `error/missing-dir` | Nonexistent dep → non-zero |
| `error/not-a-module` | Dir without go.mod → non-zero |
| `error/no-tags` | Module with no version tags → non-zero (dir mode hard) |
| `error/no-consumer-gomod` | No go.mod up walk → non-zero |
| `dry-run/no-write` | Dir mode `would: dep-update …`; no go.mod write |
| `apply/drop-replace-set-require` | Drop replace; require@latest tag; no tidy |
| `apply/nested-module-tag-prefix` | Submodule tag prefix → clean version |
| `apply/multi-dir` | Two deps updated fail-fast success |
| `all/dry-run/bumps-outdated` | Inventory pull plan: would bump + would tidy; no write |
| `all/dry-run/already-up-to-date` | Banner + summary zeros; no would: pin lines |
| `all/apply/cross-project-bump-and-tidy` | Pin + tidy; consumer only; go.sum |
| `all/apply/worktree-toplevel-not-main` | Edit linked worktree modules only, not main |
| `all/apply/skip-intra-local-replace` | Same-toplevel filesystem replace → skipped |
| `all/apply/skip-external-require` | Non-inventory require silent; inventory dep bumps |
| `all/apply/nested-owner-module-tag` | Nested owner tag prefix → clean require version |
| `all/soft/no-tag-warn` | Owner no tags → warning: + skip; exit 0 |

# Output contracts (assert targets)

## Dir mode

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

## `--all` mode

**Apply (stdout):**

```text
dep-update <module-path> -> vX.Y.Z  (tag <tag>)
go mod tidy ok  module <consumer-module-path>
dep-update: updated N, already A, skipped S
```

**Dry-run:**

```text
would: dep-update <module-path> -> vX.Y.Z
would: go mod tidy  module <consumer-module-path>
dep-update: would update N, already A, skipped S
```

**Already up to date (no pin actions)** — apply **and** dry-run use this form
(not `would update` when there are zero planned pins):

```text
dep-update: already up to date
dep-update: updated 0, already A, skipped S
```

**Summary counts:** `skipped S` = same-toplevel local filesystem replace skips +
no-tag soft-skips only. External (non-inventory) requires are silent and do
**not** increment S.

**Warnings (stderr, exit 0):** `warning:` prefix — no-tag soft skips / missing
registry paths.

**Errors (stderr, non-zero):** `wrk:` / `Error:` style consistent with existing
dep-update rejects.

Trailing `\n` after last stdout content line.

# How to Run

```sh
doctest vet ./cmd/wrk/tests/dep-update
doctest test ./cmd/wrk/tests/dep-update
```

Classic TDD: **dir-mode leaves GREEN** (including `help/mentions-flag`); new
**`all/`**, **`reject/all/`**, and **`help/mentions-all`** leaves **RED** until
implementer lands Bool + `--all` + help text.

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

	// Dir-mode fixture paths.
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

	// --all inventory fixtures.
	OwnerPath          string
	OwnerGoMod         string
	BaselineOwnerGoMod string
	MainRepo           string // consumer main when worktree leaf
	LinkedWT           string
	OwnerNestedDir     string // nested owner module dir (packages/dep)
	WantUpdated        int
	WantAlready        int
	WantSkipped        int
	WantConsumerModule string // e.g. example.com/app for tidy lines
	ProxyRoot          string
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
