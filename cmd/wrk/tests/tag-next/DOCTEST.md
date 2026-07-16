# wrk --tag-next — version tagging wire-up (Phase 3)

## Version
0.0.3

Decision tree for `wrk --tag-next`: plan and apply lightweight release tags per
tagscope scope using owned-file diffs against the latest numeric release.
Covers dry-run planning, apply (tag creation), push to origin, JSON output,
flag validation, mutual exclusion, events.jsonl, and non-git cwd errors.

# DSN (Domain Specific Notion)

- **wrk CLI** — mode `--tag-next`; invocation
  `wrk --tag-next [--dry-run] [--push] [--json] [<dir>]`. Effective cwd is
  optional `<dir>` or process cwd. May **compose** with primary `--done` /
  `--merge-back` (see `cmd/wrk/tests/done-compose/`). Still mutually exclusive
  with `--list`, `--status`, `--all-deps`, create, and other non-composed modes.
- **tagscope (dot-pkgs)** — `Plan(repoRoot, headRef)` runs Collect +
  LoadOwnedTrees + Evaluate; `Apply(repoRoot, plan, opts)` creates lightweight
  tags (`git tag <name> <head>`) and optionally pushes each new tag
  (`git push origin <tag>`). DryRun logs/plans only; Push pushes tag refs only.
- **Scope decisions** — per collected scope: skip (`-> skip` with reason) or
  plan next tag (`owned changed -> <next>`). Skip reasons include
  `prerelease-head`, `same-commit`, `no-changes`. Child scopes (e.g. `sub/`)
  evaluate independently; root skips when only child-owned paths changed.
- **Human stdout** — one line per scope decision, then apply adds
  `tagged <name> @ <short-hash>` lines; footer `N tag planned` (dry-run) or
  `N tag created` (apply). Colors when TTY/`--color` (doctest uses pipes → plain).
- **JSON stdout** — `--json` emits machine-readable plan/result on stdout (no ANSI).
- **--dry-run validation** — valid with `--all-deps`, `--tag-next`,
  `--propagate-tags`, `--sync`, and primary composition (`--done` /
  `--merge-back`); bare `wrk --dry-run` → non-zero, stderr lists those hosts.
  Primary + `--dry-run` multi-stage plans live under monotree
  `done-pipeline/dry-run/` and `merge-back-pipeline/dry-run/`.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; auto-record + events on
  every invocation.
- **events.jsonl** — successful `--tag-next` appends `command: "tag-next"`;
  composed `--tag-next --propagate-tags` still records primary `tag-next`.
- **Git fixtures** — isolated repos via `git_isolated`; `initTaggedRepo` seeds
  tags at commits and optional post-tag commits.

## Tree Overview

```
tag-next/
├── dry-run/                      # --tag-next --dry-run (plan only, no git tag)
│   ├── root-bump/                # v0.0.1 tagged, root file changed → plans v0.0.2
│   └── no-change/                # HEAD at release tag → all skip, 0 tag planned
├── apply/                        # --tag-next (create lightweight tags)
│   ├── root-bump/                # creates v0.0.2 at HEAD
│   └── skip-prerelease/          # newest v0.0.3-alpha → skip, 0 tag created
├── exclude/
│   └── sub-scope-only/           # change in sub/ only → sub/ bumps, root skips
├── json/
│   └── dry-run-root-bump/        # --tag-next --dry-run --json → JSON plan, no tag
├── push/
│   └── pushes-tag/               # bare origin + --push → tag on origin
├── flags/
│   └── dry-run-without-tag-next/ # wrk --dry-run alone → error (hosts incl. propagate-tags)
├── events/
│   ├── command-tag-next/                 # events.jsonl command=tag-next (bare)
│   └── command-tag-next-with-propagate/  # --tag-next --propagate-tags → still tag-next
└── not-git-cwd/                  # cwd not a git repo → error
```

Note: `--tag-next` + `--done`/`--merge-back` composition flag matrix lives under
`cmd/wrk/tests/done-compose/` (not mutually exclusive at flag layer).

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | dry-run/root-bump | v0.0.1 tagged, README changed → stdout plans v0.0.2; no new tag ref |
| 2 | dry-run/no-change | HEAD at v0.0.1 → skip; `0 tag planned` |
| 3 | apply/root-bump | creates lightweight tag v0.0.2 at HEAD |
| 4 | apply/skip-prerelease | v0.0.3-alpha newest → skip; no tag created |
| 5 | exclude/sub-scope-only | sub/ file changed → sub/v0.2.4; root skips |
| 6 | json/dry-run-root-bump | `--json` stdout is JSON with planned v0.0.2; no tag ref |
| 7 | push/pushes-tag | `--push` creates v0.0.2 locally and on bare origin |
| 8 | flags/dry-run-without-tag-next | bare `wrk --dry-run` → non-zero; stderr host list includes `--propagate-tags` |
| 9 | events/command-tag-next | bare success → events.jsonl `command: "tag-next"` |
| 10 | events/command-tag-next-with-propagate | `--tag-next --propagate-tags` → still `command: "tag-next"` |
| 11 | not-git-cwd | non-git cwd → non-zero, not a git repository |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/tag-next
doctest test -v ./cmd/wrk/tests/tag-next
doctest test ./cmd/wrk/tests/tag-next/dry-run/root-bump
doctest test ./cmd/wrk/tests/tag-next/apply/root-bump
doctest test ./cmd/wrk/tests/tag-next/push/pushes-tag
```

```go
import (
	"bytes"
	"os/exec"
	"testing"
)

type Request struct {
	WorkRoot   string
	WrkHome    string
	RepoDir    string   // process cwd when running wrk
	TargetDir  string   // optional first positional <dir>
	Args       []string // CLI args (e.g. --tag-next, --dry-run)
	MainRepo   string   // git repo under test (absolute)
	OriginBare string   // bare origin path for push tests
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getWrkBin(t)
	args := buildTagNextCLIArgs(req)

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = tagNextWrkEnv(req)

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