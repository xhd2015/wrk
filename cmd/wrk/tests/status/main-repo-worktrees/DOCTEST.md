# wrk --status — main-repo primary linked worktrees + optional external section

## Version
0.0.2

Decision tree for main-repo `wrk --status` when cwd resolves to a **main
repository checkout**: print **primary** paths first (main + all
`worktree.ListLinked` entries in porcelain order — including WRK out-of-tree
linked worktrees, which are **not** the external section), then—only when scan
discovers non-primary nested/dep repos—the section header
`---- external ----` (plain ASCII without color; gray `#90` / `ansiGrey` when
colorEnabled) and path-sorted **external** blocks.

# DSN (Domain Specific Notion)

- **wrk CLI** — `wrk --status` resolves the effective cwd's git toplevel. When
  the status root is a **main repo** (`worktree.IsMainRepo`), it discovers
  scan repos via `scan_repo.Scan`, lists linked worktrees via
  `worktree.ListLinked`, partitions with `PartitionStatusPaths`, and prints
  primary blocks then optional external section.
- **Dir display (`statusDirLine`)** — every `Dir:` (primary + external +
  broken/prunable) is based on **invocation cwd** (process work directory when
  wrk started), not relative-only-to-main and not always-absolute for out-of-tree
  worktrees:
  `rel = filepath.Rel(normalize(invCwd), normalize(repoPath))`; on Rel failure or
  leading `..` segment count after Clean **> 2** → absolute (`storage.NormalizePath`);
  else `filepath.ToSlash(rel)`. `.` / `../x` / `../../x` stay relative; deeper ups
  become absolute. No soft rule forcing `.` when cwd is inside the checkout.
- **Main repo checkout** — `.git` is a directory (`worktree.IsMainRepo`); enables
  primary/external sectioning.
- **Linked worktree cwd** — `.git` is a file (`worktree.IsLinked`); without
  `--main`, main-repo sectioning is skipped; output is scan-only for that cwd
  (see `from-linked-cwd`).
- **Primary membership / order** — `mainRoot` first, then every non-main path from
  `ListLinked` porcelain order (in-tree, out-of-tree WRK, dead/prunable), deduped.
  Main-owned linked paths are **primary** even when out-of-tree under
  `{WRK_HOME}/worktrees/…` — they do **not** get a section header.
- **External membership / order** — scan paths not in primary, sorted by
  normalized absolute path (nested independent repos / dep checkouts). Empty when
  only main ± its linked worktrees.
- **Section header** — when external non-empty: after last primary block, blank
  line, exactly `---- external ----` (P3: gray ANSI `ansiGrey` / `#90` when
  colorEnabled via TTY or `--color`; plain ASCII when color off), blank line,
  then external blocks. When external empty: **no** header line anywhere
  (including fixtures that only have main + its ListLinked / WRK out-of-tree
  worktrees — those stay primary).
- **Primary presentation** — main / in-scan linked / nested-in-scan use
  `printStatusBlock` (Remote/Master rules unchanged). Out-of-tree or prunable
  primary linked keep `printAppendedLinkedBlock` presentation (`Dir` via
  statusDirLine; Master for healthy; minimal for broken/prunable).
- **External presentation** — always `printStatusBlock` (no Remote for nested;
  Master only if linked).
- **Remote** — printed for the main-repo block when statusing main; gated on main
  identity, **not** on `Dir == "."`.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; external worktrees
  created via `wrk` (no args) from main repo.
- **Color** — `--color` forces ANSI on pipe; broken `error: …` value is red;
  when an external section is present, the full header line is gray (`ansiGrey` /
  `#90`) under color; plain without color. Leaves in this nested tree that have
  empty external (most cases) never print the header regardless of `--color`.

## Tree Overview

```
main-repo-worktrees/
├── no-linked-external/       # clean main only → primary [main]; no header
├── external-clean/           # main + wrk out-of-tree → primary both; no header
├── external-dirty/           # dirty out-of-tree primary linked
├── in-tree-only/             # main + in-tree linked → primary both; no header
├── mixed-external-in-tree/   # main + in-tree + out-of-tree; ListLinked order; no header
├── external-broken/          # alive path, stale gitdir → minimal error primary block
├── external-prunable/        # removed checkout → minimal prunable primary block
├── from-linked-cwd/          # --status inside external wt → no main-repo sectioning
├── ordering-two-external/    # two out-of-tree; primary order = ListLinked; no header
├── color-broken/             # --status --color → red error; still no header
├── from-main-subdir/         # cwd main/pkg/api → Dir ../.. + Remote
└── from-deep-subdir/         # cwd main/a/b/c/d → main Dir absolute + Remote
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | no-linked-external | Clean main repo; primary only; no header |
| 2 | external-clean | Main + one wrk out-of-tree wt; primary; no header; statusDirLine Dir + Master |
| 3 | external-dirty | Out-of-tree wt dirty; primary `Status: dirty (...)` |
| 4 | in-tree-only | In-tree `git worktree add` only; primary; no header |
| 5 | mixed-external-in-tree | wrk out-of-tree + in-tree wt; primary ListLinked order; no header |
| 6 | external-broken | Out-of-tree alive but git broken; `Status: error: …` |
| 7 | external-prunable | Out-of-tree checkout removed; `Status: prunable` |
| 8 | from-linked-cwd | `--status` from external wt cwd; no main-repo primary/external sections |
| 9 | ordering-two-external | Two out-of-tree wts; primary order matches ListLinked; no header |
| 10 | color-broken | `--status --color` with broken out-of-tree; red `error:`; no header |
| 11 | from-main-subdir | cwd `main/pkg/api` (≤2 ups); main Dir `../..` + Remote |
| 12 | from-deep-subdir | cwd `main/a/b/c/d` (>2 ups); main Dir absolute + Remote |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/status/main-repo-worktrees
doctest test ./cmd/wrk/tests/status/main-repo-worktrees
doctest test ./cmd/wrk/tests/status/main-repo-worktrees/external-clean
doctest test ./cmd/wrk/tests/status/main-repo-worktrees/from-main-subdir
doctest test ./cmd/wrk/tests/status/main-repo-worktrees/from-deep-subdir
```

```go
import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot      string
	WrkHome       string
	RepoDir       string // process cwd when running wrk --status
	MainRepo      string
	WtDir         string // primary external wrk worktree
	WtBranch      string
	WtDir2        string // second external wrk worktree (ordering)
	WtBranch2     string
	InTreeWtDir   string // in-tree linked worktree under main repo
	InTreeWtRel   string // relative path for scan block Dir line
	InTreeWtBranch string
	Args          []string

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product binary.
	// Prefer for leaves that do not need a process boundary. Leave false (default)
	// for true L3 e2e integration.
	InProcess bool
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	args := append([]string(nil), req.Args...)

	if req.InProcess {
		res := wrkcli.Capture(wrkcli.CaptureOpts{
			Args: args,
			Dir:  req.RepoDir,
			Env:  wrkEnv(req),
		})
		return &Response{
			Stdout:   res.Stdout,
			Stderr:   res.Stderr,
			ExitCode: res.ExitCode,
		}, nil
	}

	bin := getWrkBin(t)

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = wrkEnv(req)

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
```