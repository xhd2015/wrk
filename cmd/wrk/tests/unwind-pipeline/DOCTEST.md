# wrk --unwind composed pipeline

## Version
0.0.2

L2 decision tree for the stages that `--unwind` adds to its free-first
cross-repository peel: generated commits, land, sync, release, pin, and the
single reinstall tail. These are deliberately in-process CLI tests; fixtures
use isolated git repositories and the command itself is invoked through
`wrkcli.Capture`. Apply and failure leaves route generated commits through the
repository's mock-configured `fake-opencode` binary; they never contact a live
agent or network service.

# DSN (Domain Specific Notion)

- **Unwind coordinator** discovers dirty repositories in a checkout stack and
  peels a dependency before its consumers.
- **Peel display path** — dry-run `would: peel <path>` and apply
  `==== unwind: peel <path> ====` use the **relative path of the peel
  checkout** vs invocation cwd (same policy as status `Dir:`): nested linked
  dep → `external/dep` (or fixture-equivalent); primary at cwd → `.`.
  **Not** bare MainRepo basename alone as the full peel path. DAG identity
  remains absolute MainRepo; human pin short names remain basenames.
- **Linked worktree** may need an agent-generated commit and a land step;
  a main checkout never receives either step.
- **Gen-commit staging** — with `--gen-commit-msg`, dry-run reflects
  `--add-all` as `would: git add -A` under the peel; without `--add-all` and
  with not-fully-staged porcelain, dry-run plans
  `would: leave N file(s) uncommitted (use --add-all if necessary)`.
  Apply stages with `git add -A` **only** when `--add-all` is set (no
  unconditional pre-stage).
- **Post-land pipeline** synchronizes sibling worktrees, then releases and
  pushes the peeled repository before consumers can be pinned.
- **Dry run** is a plan only: it does not call an agent or mutate Git.

## Tree Overview

```
unwind-pipeline/
├── dry-run/full-composition/       # ordered no-mutation plan; peel display rel path
├── apply/linked-peel/              # generated commit → merge → sync → release
├── apply/multiple-linked-peels/    # one generated commit for each eligible peel
├── apply/main-only/                # no generated commit or merge for main
├── apply/add-all-stages-untracked/ # --add-all includes untracked in gen-commit
├── apply/without-add-all-no-auto-stage/  # no --add-all → untracked not pre-staged
├── failures/sync-warning/          # warning is non-fatal
├── failures/fatal-stage/           # stops later peels and reinstall tail
└── interface/
    ├── sync-accepted/              # --unwind --sync parses
    ├── incompatible-rejected/      # unrelated exclusive mode remains rejected
    └── help/                       # partners are documented
```

## Test Case Index

| Leaf | Contract | Expect |
|---|---|---|
| dry-run/full-composition | Expanded per-peel plan ordered; peel line uses `external/dep` display; Git/agent untouched | **RED** until rel peel display (and may stay GREEN on order substrings if partial) |
| apply/linked-peel | Commit precedes merge, sync, tag, and push | GREEN (stage order) |
| apply/multiple-linked-peels | Commit generation scoped per linked peel | GREEN |
| apply/main-only | Main member skips commit generation and merge-back | GREEN |
| apply/add-all-stages-untracked | `--add-all` + gen-commit includes untracked in resulting commit | **RED** until apply honors add-all only (and stages when set) |
| apply/without-add-all-no-auto-stage | without `--add-all`, untracked not auto-staged by gen-commit pre-step | **RED** until unconditional `git add -A` removed |
| failures/sync-warning | A `warning:` from sync exits successfully | GREEN |
| failures/fatal-stage | Hard stage failure prevents following peels and tail reinstall | GREEN |
| interface/* | Allowed partners, retained rejection, and help surface | GREEN |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/unwind-pipeline
doctest test ./cmd/wrk/tests/unwind-pipeline
doctest test ./cmd/wrk/tests/unwind-pipeline/dry-run/full-composition
doctest test ./cmd/wrk/tests/unwind-pipeline/apply/add-all-stages-untracked
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot        string
	WrkHome         string
	RepoDir         string
	Args            []string
	MainRepo        string
	DepMain         string
	DepWorktree     string
	SiblingWorktree string
	BeforeMain      string
	BeforeDep       string
	DoctestRoot     string
	ExtraEnv        []string
	FakeOpencode    string
	// TrackedName is a file staged before run (without-add-all leaf).
	TrackedName string
	// UntrackedName is an untracked file present at run (add-all honor leaves).
	UntrackedName string
}

type Response struct {
	Stdout, Stderr string
	ExitCode       int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = t
	_ = d
	res := wrkcli.Capture(wrkcli.CaptureOpts{Args: req.Args, Dir: req.RepoDir, WrkHome: req.WrkHome, Env: req.ExtraEnv})
	return &Response{Stdout: res.Stdout, Stderr: res.Stderr, ExitCode: res.ExitCode}, nil
}
```
