# wrk --status — append external linked worktrees from main repo

## Version
0.0.2

Decision tree for the append phase of `wrk --status`: when cwd resolves to a
**main repository checkout**, after existing `scan_repo` blocks, append one block
per **external** linked worktree not already discovered by scan.

# DSN (Domain Specific Notion)

- **wrk CLI** — `wrk --status` resolves the effective cwd's git toplevel, scans
  with `scan_repo.Scan`, and prints one status block per discovered git directory
  (unchanged scan output: relative `Dir`, optional `Master:` on in-tree linked
  worktrees only).
- **Main repo checkout** — `.git` is a directory (`worktree.IsMainRepo`); running
  `--status` from here enables the **append phase** after scan blocks.
- **Linked worktree cwd** — `.git` is a file (`worktree.IsLinked`); append phase
  is skipped; output is scan-only (same as today for that cwd).
- **Append source** — `worktree.ListLinked(mainRepo)` in `git worktree list`
  porcelain order; skip any entry whose normalized path is already in
  `scan_repo.Scan` results (dedup: in-tree linked wts like `myrepo/wt-linked`
  stay scan-only; external `{WRK_HOME}/worktrees/…` are appended).
- **Appended healthy block** — absolute normalized `Dir` (`storage.NormalizePath`);
  full fields: `Branch`, `Commit`, `Status`, `Master:` (brief branch-relation vs
  main repo current branch).
- **Appended broken block** — alive checkout path but git metadata fails: minimal
  `Dir` + `Status: error: <git stderr>` only; run continues for remaining wts.
- **Appended prunable block** — checkout dir missing (`worktree.IsDead`): minimal
  `Dir` (from porcelain) + `Status: prunable` only.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; external worktrees
  created via `wrk` (no args) from main repo.
- **Color** — `--color` forces ANSI on pipe; appended broken `error: …` value is
  red (same rule as `--projects` broken detail); prunable stays plain.

## Tree Overview

```
main-repo-worktrees/
├── no-linked-external/       # clean main, no external wt → scan only (unchanged)
├── external-clean/           # wrk external → scan `.` + appended full block
├── external-dirty/           # uncommitted change in external wt
├── in-tree-only/             # git worktree add only → no append (dedup)
├── mixed-external-in-tree/   # scan in-tree + append external only
├── external-broken/          # alive path, stale gitdir → minimal error block
├── external-prunable/        # removed checkout dir → minimal prunable block
├── from-linked-cwd/          # --status inside external wt → no append section
├── ordering-two-external/    # two external wts → append order = ListLinked
└── color-broken/             # --status --color → red error on appended block
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | no-linked-external | Clean main repo; output identical to scan-only today |
| 2 | external-clean | Main + one wrk external wt; appended full block with abs Dir + Master |
| 3 | external-dirty | External wt dirty; appended `Status: dirty (...)` |
| 4 | in-tree-only | In-tree `git worktree add` only; scan blocks only, no append |
| 5 | mixed-external-in-tree | wrk external + in-tree wt; append external only |
| 6 | external-broken | External wt alive but git broken; `Status: error: …` |
| 7 | external-prunable | External checkout removed; `Status: prunable` |
| 8 | from-linked-cwd | `--status` from external wt cwd; no appended section |
| 9 | ordering-two-external | Two external wts; append order matches `git worktree list` |
| 10 | color-broken | `--status --color` with broken external; red `error:` value |

## How to Run

```sh
doctest vet ./tests/status/main-repo-worktrees
doctest test ./tests/status/main-repo-worktrees
doctest test ./tests/status/main-repo-worktrees/external-clean
```

```go
import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
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
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getWrkBin(t)

	args := append([]string(nil), req.Args...)
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