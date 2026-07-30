# wrk --sync — CLI skeleton (P1) + IsWipSubject (P2) + full FF sync (P3)

## Version
0.0.5

Decision tree for `wrk --sync`:

- **Phase 1** — CLI skeleton: mode flag, mutual exclusion, reject positionals,
 allow `--dry-run` with `--sync`, bare `--dry-run` still rejected, stable
 zero-worktree summary on success.
- **Phase 2** — pure helper `wrkcli.IsWipSubject` (table leaves under
 `wip-subject/`; dual-mode `Run` probe).
- **Phase 3** — full FF-only bi-directional sync: pass-1 harvest (main ← each
 linked named-branch worktree when strictly FF-ahead and no WIP subjects in
 range), pass-2 distribute (each wt ← main when strictly FF-behind), dry-run
 `would:` lines, skip warnings, main-detached fatal.

Classic TDD for Phase 3: new leaves **RED** until implementer lands real
`runSync` plan+apply. Phase 1 `flags/*` and Phase 2 `wip-subject/*` ASSERT
bodies must stay GREEN (zero-summary exact stdout; probe helper).

# DSN (Domain Specific Notion)

- **wrk CLI** — standalone mode `--sync`; invocation
 `wrk --sync [--dry-run] [-v]`. No path/positional arguments when `--sync` is
 set alone. **Composable** with `--done` / `--merge-back` as a post-success
 modifier (flag order free; covered under monotree `done-sync/` and
 `merge-back-sync/`, not this nested tree). Mutually exclusive with `--list`,
 `--status`, and other non-composed mode flags (same exclusive-mode family as
 `--tag-next` for those modes).
- **Git cwd** — process cwd must be inside a git work tree (main checkout or
 linked worktree); else existing not-a-repo error. Resolve **main repo** from
 cwd. Main must be on a **named branch** (not detached) → else fatal Error.
- **Linked worktrees** — `ListLinked(main)`; skip dead paths with `warning:`.
 Only named-branch linked worktrees participate; detached wt → skip with
 warning (exit still 0).
- **Pass 1 (harvest)** — for each linked named-branch worktree: if main is
 dirty → skip that harvest with warning; if not strictly FF-ahead
 (identical / behind / diverged) → skip (diverged warns; identical/behind
 silent, not counted as skipped); if any commit subject in
 `mainBranch..wtBranch` matches `IsWipSubject` → skip with warning naming
 **first** wip short hash + subject; else
 `git -C main merge --ff-only <wtBranch>` (named branch only, never SHA).
 Re-check main after each real FF.
- **Pass 2 (distribute)** — for each linked named-branch worktree: if wt dirty
 → skip with warning; if not strictly FF-behind main (identical / ahead /
 diverged) → skip (diverged warns; identical/ahead silent); else
 `git -C wt merge --ff-only <mainBranch>`.
- **Partial skips** — exit **0**; only hard errors (e.g. main detached,
 not-a-repo, flag errors) are non-zero.
- **Skipped count** — increments only for warned skips (wip range, diverged,
 dirty main/wt, detached wt). Silent no-ops (identical / wrong-direction)
 do not increment skipped and produce no detail line. A detached linked
 worktree is skipped once (not double-counted across pass1+pass2).
- **Stdout (human, non-TTY plain)** — detail lines **only when actions > 0**,
 then a blank line, then the summary. Zero-action keeps exact one-line summary
 (Phase 1 contract):
 - real summary: `synced: %d into main, %d into worktrees, %d skipped\n`
 - dry-run summary: `would: synced: %d into main, %d into worktrees, %d skipped\n`
 - pass-1 detail: `main ← <branch> (+N commit|commits)\n`
 - pass-2 detail: `<branch> ← main (+N commit|commits)\n`
 - dry-run: every line prefixed `would: ` (details + summary)
- **Stderr warnings (plain when piped)** — examples:
 - `warning: skip <branch>: wip commit in range (<short7> <subject>)`
 - `warning: skip <branch>: diverged from main`
 - `warning: skip <branch>: dirty main`
 - `warning: skip <branch>: dirty worktree`
 - `warning: skip <path>: detached HEAD`
- **Main detached** — fatal; non-zero exit; stderr mentions detached / not on a
 named branch; no successful summary.
- **--dry-run** — plan only; `would:` stdout; **no** ref mutations
 (`git rev-parse` before == after).
- **-v** — logs major git including `merge` (existing verbose plumbing; not
 asserted in Phase 3 leaves).
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; auto-record may run
 but is not asserted in these leaves.
- **IsWipSubject** — pure helper in package `wrkcli`:
 `func IsWipSubject(subject string) bool`. After trim, case-insensitive
 prefixes `wip:`, `wip(`, `[wip]`. Empty / mid-string-only → false.
- **Dual-mode Run** — when `req.WipProbe` is true, root `Run` calls
 `wrkcli.IsWipSubject(req.Subject)` and returns `Response.IsWip` without
 invoking the CLI binary. When `req.InProcess` is true (and not WipProbe),
 root `Run` uses `wrkcli.Capture` (L2 short path) with `syncWrkEnv` +
 `buildSyncCLIArgs` (no product binary). Prefer for mutual-exclusion / early
 reject leaves. When both are false (default), product-binary CLI path.

## Tree Overview

```
sync/
├── flags/ # Phase 1 CLI skeleton (keep GREEN)
│ ├── mutual-exclusion/ # non-composable modes only (--done retired; see monotree done-sync/)
│ │ ├── with-list/
│ │ └── with-status/
│ ├── unexpected-args/
│ ├── dry-run-bare/
│ ├── dry-run-ok/ # main-only → would: zero summary (exact)
│ ├── no-linked-noop/ # main-only → zero summary (exact)
│ └── not-git/
├── wip-subject/ # Phase 2 pure IsWipSubject (keep GREEN)
│ ├── match-wip-colon/
│ ├── match-wip-colon-upper/
│ ├── match-wip-paren/
│ ├── match-bracket/
│ ├── match-bracket-upper/
│ ├── non-match-feat/
│ ├── non-match-empty/
│ └── non-match-mid/
├── pass1/ # Phase 3 — harvest main ← wt
│ ├── ff-clean/ # wt ahead, clean subjects → apply FF
│ ├── skip-wip-tip/ # tip subject is wip: → skip + warning
│ ├── skip-wip-middle/ # older unique commit is wip: → skip
│ ├── skip-diverged/
│ └── skip-dirty-main/
├── pass2/ # Phase 3 — distribute wt ← main
│ ├── ff-from-main/ # main ahead → wt FF apply
│ └── skip-dirty-wt/
├── combined/
│ └── harvest-then-distribute/ # wt1 ahead + wt2 behind → both actions
├── dry-run/
│ └── no-mutation/ # would: lines; refs unchanged
└── edge/
 ├── main-detached-error/ # fatal non-zero
 └── detached-wt-skip/ # warning skip; exit 0
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | flags/mutual-exclusion/with-list | `wrk --sync --list` → non-zero; mutually exclusive |
| 2 | flags/mutual-exclusion/with-status | `wrk --sync --status` → non-zero; mutually exclusive |
| 3 | flags/unexpected-args | `wrk --sync some-path` → non-zero; unexpected arguments |
| 4 | flags/dry-run-bare | `wrk --dry-run` → non-zero; stderr mentions `--sync` (no ``) |
| 5 | flags/dry-run-ok | `wrk --sync --dry-run` on main-only → exit 0; exact `would: synced: 0…` |
| 6 | flags/no-linked-noop | `wrk --sync` on main-only → exit 0; exact `synced: 0 into main, 0 into worktrees, 0 skipped` |
| 7 | flags/not-git | non-git cwd + `--sync` → non-zero; not a git repository |
| 8 | wip-subject/match-wip-colon | `IsWipSubject("wip: half done")` → true |
| 9 | wip-subject/match-wip-colon-upper | `IsWipSubject("WIP: foo")` → true |
| 10 | wip-subject/match-wip-paren | `IsWipSubject("wip(login): sketch")` → true |
| 11 | wip-subject/match-bracket | `IsWipSubject("[wip] experiment")` → true |
| 12 | wip-subject/match-bracket-upper | `IsWipSubject("[WIP] foo")` → true |
| 13 | wip-subject/non-match-feat | `IsWipSubject("feat: done")` → false |
| 14 | wip-subject/non-match-empty | `IsWipSubject("")` → false |
| 15 | wip-subject/non-match-mid | `IsWipSubject("chore: wip: later")` → false (not a prefix) |
| 16 | pass1/ff-clean | wt ahead +2 clean commits → main FF; detail + summary; rev-parse main==wt tip |
| 17 | pass1/skip-wip-tip | tip `wip:` → skip pass1; warning with short hash+subject; main unchanged |
| 18 | pass1/skip-wip-middle | older unique `wip:` + clean tip → still skip; names first wip |
| 19 | pass1/skip-diverged | diverged → skip; `diverged from main`; main/wt SHAs unchanged |
| 20 | pass1/skip-dirty-main | dirty main + wt ahead → skip; `dirty main`; main tip unchanged |
| 21 | pass2/ff-from-main | main ahead +1 → wt FF; detail pass2; rev-parse wt==main |
| 22 | pass2/skip-dirty-wt | main ahead + dirty wt → skip; `dirty worktree`; SHAs unchanged |
| 23 | combined/harvest-then-distribute | feature-login ahead; feature-api behind → 1 into main + 1 into worktrees |
| 24 | dry-run/no-mutation | `--dry-run` with wt ahead → `would:` details+summary; HEADs unchanged |
| 25 | edge/main-detached-error | main detached → non-zero; stderr detached / not named branch |
| 26 | edge/detached-wt-skip | linked wt detached → warning skip; exit 0; main unchanged |

Note: `--sync --done` / `--merge-back --sync` composition lives under monotree
`cmd/wrk/tests/done-sync/` and `merge-back-sync/` (retired exclusive leaf
`flags/mutual-exclusion/with-done`).

## How to Run

Most leaves are **L3 binary e2e** (`label: e2e`; costly multi-worktree leaves also `slow`).
**L2 in-process** (unlabeled): `wip-subject/*` (WipProbe) and short-path mutual-exclusion
leaves with `Request.InProcess` (`flags/mutual-exclusion/*`).
Default discovery skips labeled leaves; use `--label e2e` (or `--label-all` / `--label slow`).

```sh
doctest vet ./cmd/wrk/tests/sync
# default discovery: L2 leaves only (wip-subject + mutual-exclusion InProcess)
doctest test ./cmd/wrk/tests/sync
doctest test ./cmd/wrk/tests/sync/flags/mutual-exclusion
doctest test ./cmd/wrk/tests/sync/wip-subject
# full sync integration
doctest test --label e2e ./cmd/wrk/tests/sync
doctest test --label e2e ./cmd/wrk/tests/sync/flags/no-linked-noop
doctest test --label e2e ./cmd/wrk/tests/sync/flags/dry-run-ok
doctest test --label 'e2e && slow' ./cmd/wrk/tests/sync/pass1
doctest test --label e2e ./cmd/wrk/tests/sync/pass2
doctest test --label e2e ./cmd/wrk/tests/sync/combined
doctest test --label e2e ./cmd/wrk/tests/sync/dry-run
doctest test --label e2e ./cmd/wrk/tests/sync/edge
```

Classic TDD Phase 3: new `pass1/*`, `pass2/*`, `combined/*`, `dry-run/*`,
`edge/*` leaves **RED** until implementer lands full FF sync. Phase 1 `flags/*`
and Phase 2 `wip-subject/*` must stay GREEN (sealed ASSERT bodies unchanged).

```go
import (
	"bytes"
	"os/exec"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot string
	WrkHome string
	RepoDir string // process cwd when running wrk
	Args []string // CLI args (e.g. --sync, --dry-run)
	MainRepo string // git main repo under test (absolute), when applicable

	// Phase 2 pure-function probe (default false = CLI path for Phase 1/3 leaves).
	WipProbe bool // when true: call wrkcli.IsWipSubject(Subject); no CLI
	Subject string // commit subject for WipProbe

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product binary.
	// Prefer for mutual-exclusion / early reject leaves. Leave false for true L3 e2e.
	// Checked after WipProbe.
	InProcess bool

	// Phase 3 fixture fields (optional; set by multi-worktree Setups).
	WtPath string // primary linked worktree path
	WtBranch string // primary linked worktree branch
	WtSHA string // primary wt HEAD before run
	MainSHA string // main HEAD before run
	Wt2Path string // second linked worktree path
	Wt2Branch string // second linked worktree branch
	Wt2SHA string // second wt HEAD before run
	WipHashShort string // short=7 of first WIP commit in range (for warning pin)
	WipSubject string // full subject of first WIP commit in range
}

type Response struct {
	Stdout string
	Stderr string
	ExitCode int
	IsWip bool // set when WipProbe; result of wrkcli.IsWipSubject
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	adoptDoctestContext(d)
	if req.WipProbe {
		return &Response{
			IsWip: wrkcli.IsWipSubject(req.Subject),
		}, nil
	}

	args := buildSyncCLIArgs(req)

	if req.InProcess {
		res := wrkcli.Capture(wrkcli.CaptureOpts{
			Args: args,
			Dir: req.RepoDir,
			Env: syncWrkEnv(req),
		})
		return &Response{
			Stdout: res.Stdout,
			Stderr: res.Stderr,
			ExitCode: res.ExitCode,
		}, nil
	}

	bin := getWrkBin(t)

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = syncWrkEnv(req)

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
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```
