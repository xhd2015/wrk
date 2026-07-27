# wrk --tag-next --propagate-tags — compose pipeline (P6)

## Version
0.0.2

Decision tree for **composing** root modes `wrk --tag-next --propagate-tags`
`[--push] [--dry-run]`. Fixed stage order:

1. **tag-next** — plan/create lightweight release tags at source HEAD
2. **push** (only when `--push`) — publish **branch tip + newly created tags** to
   origin with human confirm `pushed <branch> → origin/<branch>` (same as bare
   `wrk --tag-next --push` / `runPushMain`)
3. **propagate-tags** — resolve source releases (including tags just created, or
   **planned** next tags under `--dry-run`) and bump consumers

Classic TDD for P6: production still treats the pair as mutually exclusive until
the implementer lands composition. Leaves below are the RED contract.

**Nested root:** this directory has its own `DOCTEST.md` (inheritance firewall from
parent `propagate-tags/`). No shared Setup/Run with the parent tree.

# DSN (Domain Specific Notion)

- **wrk CLI (compose)** — when both `--tag-next` and `--propagate-tags` are set
  (flag order free), run the multi-stage pipeline above instead of hard-failing
  mutual exclusion. Forms:
  - apply: `wrk --tag-next --propagate-tags`
  - dry-run: `wrk --tag-next --propagate-tags --dry-run`
  - push: `wrk --tag-next --propagate-tags --push`
  - dry-run + push: allowed at flag layer; push stage plans only (no network)
- **Stage order** — always **tag-next → push? → propagate-tags**, regardless of
  argv order. Blank line between major stdout stages (same family as
  done-pipeline multi-stage composition).
- **tag-next stage** — same as bare `wrk --tag-next` / `--dry-run` / `--push`
  (tagscope plan/apply human lines; footer `N tag planned` / `N tag created`).
- **push stage** — when `--push` and not dry-run: `runPushMain` semantics —
  branch tip + each newly created tag, plus human confirm line between major
  stages. Dry-run: `would: git push` for branch/tags; no remote mutation.
- **propagate-tags stage** — same as bare `wrk --propagate-tags` apply/dry-run,
  but source release set is:
  - **apply**: after local tags exist (including tags just created)
  - **dry-run**: **planned** next tag versions from the tag-next plan are treated
    as the would-be source releases so consumers can plan bumps to the not-yet-
    created version (core compose dry-run contract)
- **`--json` reject** — `--json` remains valid only for bare `--tag-next`. Any
  compose that includes `--propagate-tags` with `--json` → non-zero; stderr
  names `--json` and `--propagate-tags` (not silent accept, not only “mutually
  exclusive with other modes” without mentioning json).
- **Bare `--propagate-tags`** — still does **not** auto-run tag-next (unchanged;
  covered by parent `propagate-tags/` tree).
- **Out of scope (P7+)** — `--done` / `--merge-back` + propagate-tags pipeline;
  events polish for compose command identity.
- **WRK_HOME** — isolated per test; source + consumer registered in
  `projects.json`.
- **Fixtures** — source module with release tag + post-tag owned change so
  tag-next plans next patch; consumer require stuck on older version; optional
  bare origin for push; file:// module proxy for apply tidy of the next version.

## Tree Overview

```
compose/                                 # nested root (this DOCTEST.md)
├── tag-then-propagate/                  # apply: create tag then bump consumer + commit
├── dry-run/                             # plan both stages; no tag; go.mod unchanged
├── push-then-propagate/                 # --push: branch+tag on origin + confirm, then propagate
└── json-rejected/                       # --tag-next --propagate-tags --json → error
```

Split factor (MECE, significance-first):

1. **Outcome class** — success apply | success dry-run plan | hard reject (json).
2. Within apply: **push stage** present or absent (siblings: tag-then-propagate vs
   push-then-propagate).

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| C1 | tag-then-propagate | `--tag-next --propagate-tags`: source gets next tag; app require bumps to that version; consumer commit with `chore(deps):` |
| C2 | dry-run | `--tag-next --propagate-tags --dry-run`: both stages planned; no new tag; consumer go.mod/HEAD unchanged |
| C3 | push-then-propagate | `--tag-next --propagate-tags --push`: branch+tag on origin; confirm line; then consumer bump+commit |
| C4 | json-rejected | `--tag-next --propagate-tags --json` → non-zero; stderr mentions json + propagate-tags |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/propagate-tags/compose
doctest test -v ./cmd/wrk/tests/propagate-tags/compose
doctest test ./cmd/wrk/tests/propagate-tags/compose/tag-then-propagate
doctest test ./cmd/wrk/tests/propagate-tags/compose/dry-run
doctest test ./cmd/wrk/tests/propagate-tags/compose/push-then-propagate
doctest test ./cmd/wrk/tests/propagate-tags/compose/json-rejected
```

All four leaves are **Classic TDD RED** until P6 composition is implemented
(today: `--tag-next` + `--propagate-tags` → mutual exclusion at flag layer).

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
	WrkHome  string
	RepoDir  string // process cwd when running wrk (source main)
	Args     []string

	// Fixture paths filled by leaves / shared helpers.
	SourcePath string // source main repo (lib)
	AppPath    string // consumer project
	OriginBare string // bare origin for push leaves

	// Expected next release from tag-next (fixture contract).
	OldTag     string // e.g. v1.0.0
	NextTag    string // e.g. v1.0.1
	ModulePath string // e.g. example.com/lib

	// Pre-run snapshots for non-mutation asserts.
	AppGoModBefore    string
	SourceGoModBefore string
	SourceHEADBefore  string
	AppHEADBefore     string
	SourceTagsBefore  string
	AppTagsBefore     string

	// Optional env for Run (apply leaves set file:// GOPROXY for tidy).
	ExtraEnv []string

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product binary.
	// Prefer for flag-layer reject leaves (e.g. json-rejected) that do not need a process boundary.
	// Leave false (default) for true L3 e2e integration.
	InProcess bool
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	adoptDoctestContext(d)
	args := append([]string(nil), req.Args...)

	if req.InProcess {
		res := wrkcli.Capture(wrkcli.CaptureOpts{
			Args: args,
			Dir:  req.RepoDir,
			Env:  composeWrkEnv(req),
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
	cmd.Env = composeWrkEnv(req)

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
