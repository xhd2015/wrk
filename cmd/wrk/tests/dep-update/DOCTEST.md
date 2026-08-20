# wrk --dep-update — dir pin + inventory pull (`--all`)

## Version
0.0.3

Decision tree for **`wrk --dep-update`** in two modes:

1. **Dir mode**: `wrk --dep-update <dir>… [--dry-run]` — resolve module path +
   latest tag from each `<dep-dir>` (`update.Pin`). Consumer set =
   **`CollectStackInventory(cwd)`** (cwd git toplevel + nested status repos +
   BFS over local filesystem replaces). Scan **every** `go.mod` under every
   member `Path` and **pin every module that already requires** the path (do
   not add new requires). Self (`consumer.Path == dep.Path`) never pinned.
   Not git → today’s fallback (nearest `go.mod`, already-require). After pins:
   **versioned `go mod tidy`** via `gotool/withgo` unless `vendor/` sits beside
   that `go.mod` (once per affected module after all its pins). Multi-arg
   **validate every dir first** (dry-run: any bad arg → no banner / no tree).
   Zero requirers on the whole stack → `wrk:` error containing `requires` (no
   banner). wrk does **not** exec kool; never `go mod vendor`.
2. **`--all` inventory pull**: `wrk --dep-update --all [--dry-run]` — same
   **stack consumer set**; match requires against `BuildInventory(WRK_HOME)`,
   pin inventory-owned deps to latest tags, then the **same tidy helper**
   (versioned + vendor skip) once per affected consumer. `--all` still requires
   a git repo.

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
| D2 | After pins: versioned `go mod tidy` (`withgo.ModuleGoLine` + `withgo.Run`) unless `<modDir>/vendor` exists → `skip tidy  module …  (vendor/)`. Never `go mod vendor`. Missing go line → fail |
| D3 | Multi-arg **fail-fast** |
| D4 | Dir mode does not use inventory / `--all` |
| D5 | Structured CLI tree (`====`, `dep`, `checkout`, `module`, `pin` / `would: pin`, tidy / skip tidy); no kool commit-message line; wrk does **not** exec kool |
| D6 | Consumer set = `CollectStackInventory(cwd)` when git; else nearest `go.mod`. Scan every `go.mod` under every member `Path`. Pin only modules that **already require** the path. Self never pinned |
| D7 | Pin does not add new requires; zero matching consumers on the whole stack → `wrk:` error containing `requires`; **no banner** |
| D8 | One banner per invocation. Deps named once at top (argv order). Body is checkout → module → actions. Tidy once under the module. Dir-mode summary `updated N modules in C checkouts` |
| D9 | Consumer whose dep is covered by a **same-toplevel local filesystem replace** (intra-module replace) is **skipped**, not pinned. Prints `skip  <dep>  (intra-module replace)` (dry-run: `would: skip …`); no tidy for that module. Summary gains `, skipped S` when S > 0. Mirrors `--all` A4 |

**Locked product rules (`--all`):**

| ID | Rule |
|----|------|
| A1 | Consumer set = **`CollectStackInventory(cwd)`** (not only cwd git toplevel; still not main-repo when cwd is a linked worktree Path) |
| A2 | Scan all go.mod under every stack member `Path`; mutate **only** those modules |
| A3 | Inventory read-only (`BuildInventory`); ownership + latest tags from owner main path |
| A4 | External require → **silent skip** (does **not** increment `skipped S`); same-toplevel + filesystem replace → **skipped** (counts in S; no bump) |
| A5 | No tag on owner → `warning:` stderr + skip (counts in S); exit 0 |
| A6 | Already at latest → **already** count (no per-dep noise) |

| A7 | After apply pins: **same tidy helper as dir-mode** (versioned `withgo` + vendor skip) once per consumer with ≥1 pin; no commit/build |
| A8 | Dry-run: validate first; then banner `==== dep-update (dry-run) ====`, `would:` action lines, summary `would update N, already A, skipped S in C checkouts`; zero writes |
| A9 | Consumer need **not** be registered (pull-only) |

**Partners:** `--dry-run`, optional `--color` / `--no-color`, `-v`.

**Out of scope:** `--dep-replace --all` / `--dep-replace` fan-out; commit/build
gate; other wrk tidy sites (bring, pin-locals, unwind, propagate-tags); JSON;
editing other projects' go.mod; kool CLI; real network / real SDK download.

# DSN (Domain Specific Notion)

- **wrk --dep-update** — exclusive top-level Bool mode. Dir args (remaining) **or**
  partner **`--all`**. Partners: `--dry-run`, color, `-v`. Does not exec kool.
- **Stack consumer set** — `CollectStackInventory(cwd)`: cwd git toplevel +
  nested status repos + BFS over local filesystem replaces (same as `--unwind` /
  `--pin-locals`). Scan every `go.mod` under every member `Path`. Not git →
  nearest `go.mod`. `--all` still requires a git repo.
- **Dir mode** — for each dep dir: resolve module + latest tag (`update.Pin`).
  Pin every stack module that already requires that path. Self never pinned.
  Same-toplevel local filesystem replace (intra-module replace) → **skip**
  (D9): printed as `skip  <dep>  (intra-module replace)`, no tidy. Then
  versioned tidy or `skip tidy  (vendor/)` once per affected module with pins.
  Zero requirers on the stack → `wrk:` error containing `requires`; no banner.
- **`--all` mode** — same stack consumer set; for each require, consult inventory
  ownership and owner tags; Pin when outdated inventory-owned; external require
  silent (not in `skipped S`); same-checkout filesystem replace + no-tag
  soft-skips count in `skipped S`; same tidy helper after apply.
- **Inventory** — `BuildInventory(WRK_HOME)`: registered projects + sub-modules;
  latest numeric tags from owner on-disk module dir (main path).
- **withgo tidy** — `ModuleGoLine` + `Run(ver, ["go","mod","tidy"], WithGo, …)`.
  Pin `go 1.22` → `go1.22.12`. Skip when `vendor/` is a directory beside go.mod.
- **CLI tree** — one banner; dir-mode `dep` headers (argv order); body
  checkout → module → `pin` / `would: pin` + one tidy. Checkout label =
  `statusDirLine` vs invocation cwd (`.` / `external/kool`).
- **Dry-run** — validate every dir arg first (any bad → no banner / no tree);
  then `would:` action lines; zero writes.
- **Hard errors** — empty mode (no dirs, no `--all`); `--all` with paths; bare
  `--all`; exclusive modes; dir-mode missing dir / not module / no tags / no
  consumer go.mod / zero requirers; Pin/tidy write failures (fail-fast). No banner.
- **Soft warnings (`--all`)** — missing registry paths, no-tag skips (`warning:`).

# Decision Tree

```text
dep-update/
├── help/
│   ├── mentions-flag                # --dep-update
│   ├── mentions-all                 # --all with --dep-update
│   └── mentions-stack               # unwind/stack + --dry-run
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
│   ├── no-consumer-gomod
│   └── no-consumer-requires         # stack has no requirer of xxx
├── dry-run/                         # dir mode
│   ├── no-write
│   ├── vendor-skip
│   ├── stack-no-write
│   ├── multi-dir-stack-no-write
│   ├── skip-intra-local-replace     # would: skip on dep's sub-module
│   └── bad-second-arg               # no banner; first dep not a half-plan
├── apply/                           # dir mode
│   ├── drop-replace-set-require     # pin + tidy (nearest / not-git)
│   ├── nested-module-tag-prefix
│   ├── multi-dir                    # two pins + tidy once (one consumer)
│   ├── fan-out-requirers            # same-checkout: root + pkg/ both require
│   ├── skip-non-requirer            # same-checkout sibling, default quiet
│   ├── vendor-skip
│   ├── versioned-tidy               # go 1.19 → go1.19.13 wrapper
│   ├── stack-requirer-other-checkout
│   ├── stack-skip-non-requirer      # cross-checkout, default quiet
│   ├── stack-skip-self
│   ├── skip-intra-local-replace      # dep's own sub-module with replace => ../ skipped
│   └── multi-dir-stack
└── all/
    ├── dry-run/
    │   ├── bumps-outdated
    │   ├── already-up-to-date
    │   └── stack-outdated
    ├── apply/
    │   ├── cross-project-bump-and-tidy
    │   ├── worktree-toplevel-not-main
    │   ├── skip-intra-local-replace
    │   ├── skip-external-require
    │   ├── nested-owner-module-tag
    │   ├── vendor-skip
    │   └── stack-outdated
    └── soft/
        └── no-tag-warn
```

# Test Index

| Path | Intent |
|------|--------|
| `help/mentions-flag` | Root help mentions `--dep-update` |
| `help/mentions-all` | Root help mentions `--all` in context of `--dep-update` |
| `help/mentions-stack` | Root help mentions unwind/stack (or equivalent) + `--dry-run` for `--dep-update` |
| `reject/no-args` | Neither dirs nor `--all` → requires directory or `--all` |
| `reject/with-dep-replace` | XOR with `--dep-replace` |
| `reject/with-pin-locals` | Exclusive with `--pin-locals` |
| `reject/all/bare-all` | Bare `--all` without `--dep-update` → error |
| `reject/all/all-with-dirs` | `--dep-update --all` with path args → error |
| `error/missing-dir` | Nonexistent dep → non-zero |
| `error/not-a-module` | Dir without go.mod → non-zero |
| `error/no-tags` | Module with no version tags → non-zero (dir mode hard) |
| `error/no-consumer-gomod` | No git + no go.mod ancestor → non-zero |
| `error/no-consumer-requires` | Stack has no requirer of xxx → `requires` error; no banner |
| `dry-run/no-write` | Dry-run tree (`would: pin` + `would: go mod tidy`); no write |
| `dry-run/vendor-skip` | Dry-run tree `would: skip tidy  (vendor/)`; no write |
| `dry-run/stack-no-write` | Single-target stack tree; go.mod/go.sum unchanged |
| `dry-run/multi-dir-stack-no-write` | Two `dep` headers; would: pins; no write |
| `dry-run/skip-intra-local-replace` | `would: skip` on dep's sub-module with intra-module replace; no write |
| `dry-run/bad-second-arg` | No banner; `wrk:` + missing dir; first dep not a half-plan |
| `apply/drop-replace-set-require` | Drop replace; require@latest; tree + tidy; go.sum |
| `apply/nested-module-tag-prefix` | Submodule tag prefix → clean version + tidy |
| `apply/multi-dir` | Two pins + tidy once for the one consumer |
| `apply/fan-out-requirers` | Same-checkout root + `pkg/` both require xxx → both pinned + tidied |
| `apply/skip-non-requirer` | Same-checkout sibling without require unchanged; default quiet |
| `apply/vendor-skip` | Pin applied; `skip tidy  (vendor/)`; no go.sum |
| `apply/versioned-tidy` | Consumer `go 1.19`; wrapper at `go1.19.13` used |
| `apply/stack-requirer-other-checkout` | Pin+tidy primary **and** other git checkout that requires xxx |
| `apply/stack-skip-non-requirer` | Other-checkout module without require unchanged; default quiet |
| `apply/stack-skip-self` | Dep’s own go.mod not pinned when dep checkout is on the stack |
| `apply/multi-dir-stack` | Two dep args; one consumer both pins + one tidy; other requires only first |
| `apply/skip-intra-local-replace` | Dep's sub-module with intra-module replace `=> ../` skipped; primary pinned |
| `all/dry-run/bumps-outdated` | `--all` dry-run tree; no argv `dep` list; no write |
| `all/dry-run/already-up-to-date` | Banner + zero summary `in C checkouts`; no pin tree |
| `all/dry-run/stack-outdated` | Dry-run tree for other-checkout inventory require; no argv `dep` list |
| `all/apply/cross-project-bump-and-tidy` | Pin + tidy tree; consumer only; go.sum |
| `all/apply/worktree-toplevel-not-main` | Edit linked worktree Path only, not MainRepo |
| `all/apply/skip-intra-local-replace` | Same-checkout filesystem replace → skipped |
| `all/apply/skip-external-require` | Non-inventory require silent; inventory dep bumps |
| `all/apply/nested-owner-module-tag` | Nested owner tag prefix → clean require version |
| `all/apply/vendor-skip` | `--all` pin + skip tidy when vendor/ present |
| `all/apply/stack-outdated` | Inventory-owned require on **other** stack checkout pinned + tidied |
| `all/soft/no-tag-warn` | Owner no tags → warning: + skip; exit 0 |

# Output contracts (assert targets)

Tokens: `====`, `dep`, `checkout`, `module`, `pin`, `would:`,
`go mod tidy ok`, `skip tidy`, `skip`, `(vendor/)`, `(intra-module replace)`, `->`.
Trailing `\n`.
Checkout path = `statusDirLine` vs invocation cwd (`.` / `external/kool`).
No short form for a single target. Default quiet: only modules that change
or are skipped (`-v` would list `no require` / `self` / `already <ver>` — optional).

## Dir mode

**Apply success (stdout, trailing `\n`) — banner, argv `dep` headers, checkout → module → pin + tidy:**

```text
==== dep-update ====
dep  <dep-path> -> <new>  (tag <tag>)

  checkout  .
    module  <consumer-path>
      pin  <dep-path>  <old> -> <new>
      go mod tidy ok

dep-update: updated N modules in C checkouts
```

Tag parenthetical on the header `dep` line is optional if no tag; if present,
`(tag …)`. Pin lines include **old -> new**. Vendor: `skip tidy  (vendor/)`
under that module (indent with the tree).

**Multiple targets:** N `dep` header lines (argv order). Each module lists
only pins for deps it already requires, then one tidy.

**Intra-module replace skip (D9):** a consumer whose dep is covered by a
same-toplevel local filesystem replace (intra-module replace) is **skipped**,
not pinned. Under that module: `skip  <dep>  (intra-module replace)` (dry-run:
`would: skip  <dep>  (intra-module replace)`). No tidy for skip-only modules.
When skips exist, the summary gains `, skipped S`:
`dep-update: updated N modules, skipped S in C checkouts`
(dry-run: `would update N modules, skipped S in C checkouts`).

**Dry-run:** banner `==== dep-update (dry-run) ====`; `would: pin  …`;
`would: go mod tidy` or `would: skip tidy  (vendor/)`;
`would: skip  …  (intra-module replace)`;
summary `dep-update: would update N modules[, skipped S] in C checkouts`. Zero writes.

**Zero requirers (stderr, non-zero):** `wrk:` error containing `requires`.
**No banner.**

**Bad second arg on dry-run:** `wrk:` + `no such dir` (or equivalent); no
banner / no tree; first dep not described as a half-plan.

## `--all` mode

Same banner + checkout/module tree; **no** argv `dep` header list
(inventory chooses).

**Apply (stdout):**

```text
==== dep-update ====

  checkout  .
    module  <consumer-path>
      pin  <dep-path>  <old> -> <new>
      go mod tidy ok

dep-update: updated N, already A, skipped S in C checkouts
```

**Dry-run:** banner `==== dep-update (dry-run) ====`; `would: pin` /
`would: go mod tidy`; summary
`dep-update: would update N, already A, skipped S in C checkouts`.

**Already up to date (no pin actions)** — keep a banner + zero summary;
no pin tree required. Apply **and** dry-run use apply wording
(not `would update` when there are zero planned pins):

```text
==== dep-update ====
dep-update: already up to date
dep-update: updated 0, already A, skipped S in C checkouts
```

**Summary counts:** `skipped S` = same-checkout local filesystem replace skips +
no-tag soft-skips only. External (non-inventory) requires are silent and do
**not** increment S.

**Warnings (stderr, exit 0):** `warning:` prefix — no-tag soft skips / missing
registry paths.

**Errors (stderr, non-zero):** `wrk:` / `Error:` style consistent with existing
dep-update rejects. **No banner.**

Trailing `\n` after last stdout content line.

# How to Run

```sh
doctest vet ./cmd/wrk/tests/dep-update
doctest test ./cmd/wrk/tests/dep-update
```

Classic TDD: **new stack / new-stdout** leaves and **rewritten** dir-mode /
`--all` success/dry-run tree asserts stay **RED** until the implementer lands
stack fan-out + CLI tree. Reject/error leaves keep today’s meaning and may stay
**GREEN**. `help/mentions-stack` is RED until help mentions unwind/stack.

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
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
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
	WantCheckouts      int    // C in "in C checkouts"; 0 → helper default 1
	WantOldVersion     string // pin old version (dir-mode v0.0.1 / --all v1.0.0)
	WantOldVersion2    string
	WantCheckout       string // statusDirLine vs cwd; default "."
	WantCheckout2      string // other stack checkout, e.g. external/kool
	WantConsumerModule string // e.g. example.com/app for module lines
	ProxyRoot          string

	// Versioned tidy seam: leaves seed $InstallDir/<pin>/bin/go wrappers.
	InstallDir    string
	WithGo        withgo.ResolveOptions
	WrapperRecord string // dest last-run file written by the go wrapper
	WantGoPin     string // e.g. go1.19.13

	// Extra consumer (fan-out / skip-non-requirer sibling).
	Consumer2ModDir     string
	Consumer2GoMod      string
	Baseline2GoMod      string
	WantConsumer2Module string
	VendorDir           string
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
			Args:   args,
			Dir:    dir,
			Env:    depUpdateEnv(req),
			WithGo: req.WithGo,
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
