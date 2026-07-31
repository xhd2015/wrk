# wrk --unwind composed pipeline

## Version
0.0.2

L2 decision tree for the stages that `--unwind` adds to its free-first
cross-repository peel: generated commits, land, sync, release, pin, and the
single reinstall tail.  These are deliberately in-process CLI tests; fixtures
use isolated git repositories and the command itself is invoked through
`wrkcli.Capture`. Apply and failure leaves route generated commits through the
repository's mock-configured `fake-opencode` binary; they never contact a live
agent or network service.

# DSN (Domain Specific Notion)

- **Unwind coordinator** discovers dirty repositories in a checkout stack and
  peels a dependency before its consumers.
- **Linked worktree** may need an agent-generated commit and a land step;
  a main checkout never receives either step.
- **Post-land pipeline** synchronizes sibling worktrees, then releases and
  pushes the peeled repository before consumers can be pinned.
- **Dry run** is a plan only: it does not call an agent or mutate Git.

## Tree Overview

```
unwind-pipeline/
├── dry-run/full-composition/       # accepted, ordered no-mutation plan
├── apply/linked-peel/              # generated commit → merge → sync → release
├── apply/multiple-linked-peels/    # one generated commit for each eligible peel
├── apply/main-only/                # no generated commit or merge for main
├── failures/sync-warning/          # warning is non-fatal
├── failures/fatal-stage/           # stops later peels and reinstall tail
└── interface/
    ├── sync-accepted/              # --unwind --sync parses
    ├── incompatible-rejected/      # unrelated exclusive mode remains rejected
    └── help/                       # partners are documented
```

## Test Case Index

| Leaf | Contract |
|---|---|
| dry-run/full-composition | Expanded per-peel plan is ordered and Git/agent remain untouched. |
| apply/linked-peel | Commit precedes merge, sync, tag, and push. |
| apply/multiple-linked-peels | Commit generation is scoped to each linked peel. |
| apply/main-only | A main member skips commit generation and merge-back. |
| failures/sync-warning | A `warning:` from sync exits successfully. |
| failures/fatal-stage | A hard stage failure prevents following peels and tail reinstall. |
| interface/* | Allowed partners, retained rejection, and help surface. |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/unwind-pipeline
doctest test ./cmd/wrk/tests/unwind-pipeline
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot string
	WrkHome string
	RepoDir string
	Args []string
	MainRepo string
	DepMain string
	DepWorktree string
	SiblingWorktree string
	BeforeMain string
	BeforeDep string
	DoctestRoot string
	ExtraEnv []string
	FakeOpencode string
}

type Response struct { Stdout, Stderr string; ExitCode int }

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = t
	_ = d
	res := wrkcli.Capture(wrkcli.CaptureOpts{Args: req.Args, Dir: req.RepoDir, WrkHome: req.WrkHome, Env: req.ExtraEnv})
	return &Response{Stdout: res.Stdout, Stderr: res.Stderr, ExitCode: res.ExitCode}, nil
}
```
