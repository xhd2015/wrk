# wrk --propagate-tags — plan and apply consumer tag updates (P3–P5)

## Version

**Layer: L2 in-process CLI** via `wrkcli.RunCLI`.
0.0.2

Decision tree for root-level exclusive mode `wrk --propagate-tags`. From the
source project's git main repo (cwd), resolve latest numeric release tags via P1
`ResolveSourceReleases`, scan **other** registered projects for matching
`require` lines, and either **plan** (`--dry-run`) or **apply** (no dry-run)
consumer go.mod updates to those versions.

**P3 dry-run** is GREEN. **P4 apply** (drop replace + require + tidy) is GREEN.
**P5 build gate + commit** (Classic TDD RED): after tidy, run `go build ./...`
per updated consumer module; if all touched modules in a consumer project build
OK → one git commit (go.mod/go.sum only) with `chore(deps):` subject; if build
fails → stderr `warning:`, leave dirty tree, no commit, continue (exit 0).

# DSN (Domain Specific Notion)

- **wrk CLI** — top-level exclusive mode `--propagate-tags`. Forms:
  - plan: `wrk --propagate-tags --dry-run`
  - apply: `wrk --propagate-tags` (no `--dry-run`)
  From a git work tree. Source = main repo of cwd (same main-repo resolution
  family as other git modes). Mutually exclusive with `--list` and other mode
  flags (same family as `--projects` / `--tag-next`).
- **Source releases (P1)** — `ResolveSourceReleases(sourceMain)` maps each scanned
  source module to `{ModulePath, Tag, Version}` for the latest **numeric** release
  tag (root `vX.Y.Z`, nested `sub/vX.Y.Z` → version `vX.Y.Z`). Modules without a
  numeric tag are not plan-ready; when the source has **no** usable release tags
  for its modules, the mode **hard-errors** (exit ≠ 0).
- **Consumer discovery** — `storage.ListProjects(WRK_HOME)` other than the source
  main repo. For each other project, scan Go modules; select when a `require`
  module path matches a source release **and** the require version **differs**
  from the release version. Intra-project / self-project modules are skipped
  (soft-warn if encountered as consumers). Unknown/external requires are ignored.
- **Local replace (cross-project)** — when a consumer module has a local
  `replace` for a source module that would be updated:
  - dry-run: plan includes `would: drop replace …` (no write)
  - apply: **drop** the replace via `go mod edit -dropreplace`, then bump require
- **Apply edits (P4)** — per updated consumer module: drop cross-project local
  replace(s), `go mod edit -require=mod@version` to the source release version,
  then `go mod tidy`.
- **Build gate + commit (P5)** — after tidy, per updated consumer **module dir**:
  `go build ./...`. If **all** touched modules in a consumer **project** build OK
  → one git commit on that project staging **only** edited `go.mod` / `go.sum`
  under those module dirs; subject e.g.
  `chore(deps): bump example.com/lib to v1.2.3`. If any touched module in the
  project fails build → print `warning:` on stderr, leave the working tree dirty
  (bumped require retained), **no** commit for that project, continue other
  projects; overall exit **0** (partial success). Dry-run does **not** print
  `would: build` / `would: commit` (P3 plan stays action-plan only).
- **Human stdout (plan, dry-run)** — pipes → plain text (no ANSI):

  ```text
  source: <abs-source-main>
    <module-path>  @ <version>  (tag <tag>)

  would: update <consumer-module>  (project <basename>)
    <dep-module>  <old-version> -> <new-version>

  would: drop replace <dep-module>  (project <basename>)

  would: update N module(s) across M project(s)
  ```

- **Human stdout (apply)** — same layout, past-tense verbs, **no** `would:` prefix.
  On successful build+commit for a consumer update, **additive** indented lines
  after the version arrows (and before any drop-replace block / next section):

  ```text
  source: <abs-source-main>
    <module-path>  @ <version>  (tag <tag>)

  updated <consumer-module>  (project <basename>)
    <dep-module>  <old-version> -> <new-version>
    go build ./... ok
    committed <short7>  chore(deps): bump <dep-module> to <new-version>

  dropped replace <dep-module>  (project <basename>)

  updated N module(s) across M project(s)
  ```

  - `source:` block lists every resolved source release (root + nested).
  - One consumer block per module that needs at least one bump; indented lines
    show each dep arrow.
  - After bumps, when build succeeds: `  go build ./... ok`.
  - When the project is committed: `  committed <short7>  <subject>` where
    `<short7>` is `git rev-parse --short=7 HEAD` after the commit and subject
    matches `chore(deps): bump <module> to <version>` (single-dep form used by
    these leaves; multi-dep may summarize).
  - When build fails: **no** `go build ./... ok` and **no** `committed` line for
    that project; version arrows still shown; footer still counts the module
    update (go.mod was edited).
  - Drop-replace lines only when a local replace is (or would be) removed;
    still reported after the update block as in P3/P4.
  - Footer always present with English pluralization; ends with `\n`.
  - Already-current consumers produce **no** consumer module block; footer zeros;
    source block still shown when releases exist; **no** build/commit.
- **stderr** — soft issues use `warning:` prefix (e.g. missing registry path,
  **build failure** leaving changes uncommitted). Hard errors use `wrk:` / clear
  phrasing; stdout empty on hard mutual-exclusion and invalid dry-run host paths
  when applicable.
- **Side effects (dry-run)** — after a successful plan: consumer/source `go.mod`
  bytes unchanged; no new/removed git tags; git HEAD unchanged; no commits.
- **Side effects (apply, build OK)** — consumer `go.mod` (and typically `go.sum`)
  updated and **committed** on the consumer project; source `go.mod` / tags /
  HEAD unchanged; consumer tags unchanged; consumer HEAD advances by one commit
  whose tree only touches go.mod/go.sum under edited modules.
- **Side effects (apply, build fail)** — consumer `go.mod` may be dirty at the
  bumped require; **no** new commit (HEAD unchanged); source unchanged; stderr
  `warning:`.
- **--dry-run host list** — bare `wrk --dry-run` remains invalid; stderr must list
  valid hosts **including** `--propagate-tags` (alongside existing `--done`,
  `--merge-back`, `--all-deps`, `--tag-next`, `--sync`).
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; projects registered via
  seeded `projects.json`.
- **events.jsonl** — successful bare `--propagate-tags` appends
  `command: "propagate-tags"`. Compose with `--tag-next` is primary `tag-next`
  (covered under `tag-next/events/`).
- **Fixtures / tidy** — apply leaves may seed a local `file://` module proxy so
  `go mod tidy` can resolve synthetic `example.com/*` release versions offline.
- **Compose with `--tag-next` (P6)** — nested root `compose/` (own `DOCTEST.md`
  firewall): `wrk --tag-next --propagate-tags [--push] [--dry-run]`. Not inherited
  by this tree; run `doctest test ./cmd/wrk/tests/propagate-tags/compose`.

## Tree Overview

```
propagate-tags/
├── dry-run/                                 # --propagate-tags --dry-run success plans
│   ├── would-update/
│   │   ├── root-and-sub/                    # source root+sub tags; consumer outdated
│   │   └── with-local-replace/              # plan would drop replace; go.mod unchanged
│   ├── already-current/                     # consumer require == release; no would-update
│   └── no-matching-consumer/                # other project does not require source
├── apply/                                   # --propagate-tags (no --dry-run) real edits
│   ├── root-bump/                           # require bumps + build ok + commit (P5)
│   ├── drop-replace-and-bump/               # drop local replace + bump + build + commit
│   ├── build-ok-commits/                    # P5 focus: HEAD moves; chore(deps) subject
│   ├── build-fail-no-commit/                # P5: compile fail → warning:; no commit
│   └── already-current/                     # exit 0; go.mod unchanged; no build/commit
├── errors/
│   ├── not-git-cwd/                         # non-git cwd → error
│   ├── no-source-tags/                      # source modules lack numeric tags → hard error
│   └── mutual-exclusion/
│       └── with-list/                       # + --list → exclusive error
├── flags/
│   └── dry-run-without-propagate-tags/      # bare --dry-run host list includes --propagate-tags
├── events/
│   └── command-propagate-tags/              # events.jsonl command=propagate-tags
└── compose/                                 # nested DOCTEST root (P6) — see compose/DOCTEST.md
    ├── tag-then-propagate/
    ├── dry-run/
    ├── push-then-propagate/
    └── json-rejected/
```

Split factor (MECE, significance-first):

1. **Invocation class** — dry-run plan | apply | hard errors | dry-run host flag validation | events.
2. Within dry-run: **consumer plan outcome** (outdated require | already current |
   no matching require).
3. Within dry-run would-update: **replace overlay** (plain require bump | local replace plan).
4. Within apply: **consumer apply outcome** — edit shape (root bump | drop-replace+bump),
   **build gate** (ok→commit | fail→no-commit), already-current no-op.
5. Within errors: precondition failure kind (not git | no tags | mutual exclusion).

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| D1 | dry-run/would-update/root-and-sub | lib root+sub releases; app requires older → source block + would-update lines; no mutation |
| D2 | dry-run/would-update/with-local-replace | outdated require + local replace → `would: drop replace`; go.mod/tags/HEAD unchanged |
| D3 | dry-run/already-current | app already at release versions → no would-update module block; footer 0 |
| D4 | dry-run/no-matching-consumer | other registered project does not require source modules → footer 0 |
| A1 | apply/root-bump | non-dry-run bumps require; stdout `updated` + `go build ./... ok` + `committed`; HEAD advances |
| A2 | apply/drop-replace-and-bump | drops cross-project local replace + sets require; build ok + commit |
| A3 | apply/already-current | apply no-op: exit 0, go.mod unchanged, `updated 0 modules…`; no commit |
| A4 | apply/build-ok-commits | P5: clean fixture; new commit with `chore(deps):` subject; go.mod at target; only mod files |
| A5 | apply/build-fail-no-commit | P5: consumer fails `go build ./...`; stderr `warning:`; HEAD unchanged; go.mod dirty at bump |
| E1 | errors/not-git-cwd | non-git cwd → exit ≠ 0; stderr not a git repository |
| E2 | errors/no-source-tags | source has modules but no numeric tags → exit ≠ 0 |
| E3 | errors/mutual-exclusion/with-list | `--propagate-tags --list` → exit ≠ 0; mutually exclusive |
| F1 | flags/dry-run-without-propagate-tags | bare `wrk --dry-run` → non-zero; stderr lists hosts including `--propagate-tags` |
| V1 | events/command-propagate-tags | bare success → events.jsonl `command: "propagate-tags"` |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/propagate-tags
doctest test -v ./cmd/wrk/tests/propagate-tags
doctest test ./cmd/wrk/tests/propagate-tags/dry-run/would-update/root-and-sub
doctest test ./cmd/wrk/tests/propagate-tags/apply/root-bump
doctest test ./cmd/wrk/tests/propagate-tags/apply/build-ok-commits
doctest test ./cmd/wrk/tests/propagate-tags/apply/build-fail-no-commit
doctest test ./cmd/wrk/tests/propagate-tags/errors/no-source-tags
doctest test ./cmd/wrk/tests/propagate-tags/flags/dry-run-without-propagate-tags
```

Dry-run + errors + flags (D*, E*, F1) and apply/already-current stay GREEN.
Apply leaves that expect build+commit (A1, A2, A4) and build-fail (A5) are **RED**
until implementer lands P5 (`go build ./...` gate + optional commit).

```go
import (
	"github.com/xhd2015/wrk/wrkcli"
	"strings"
	"bytes"
	"os/exec"
	"testing"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // process cwd when running wrk
	Args     []string

	// Fixture paths filled by leaves for Assert templates / side effects.
	SourcePath string // source main repo (cwd for success leaves)
	AppPath    string // consumer project (when present)
	OtherPath  string // non-consumer other project (when present)

	// Pre-run snapshots for non-mutation asserts.
	AppGoModBefore    string
	SourceGoModBefore string
	SourceHEADBefore  string
	AppHEADBefore     string
	SourceTagsBefore  string
	AppTagsBefore     string

	// Optional env for Run (apply leaves may set file:// GOPROXY for tidy).
	ExtraEnv []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}


func wrkDateForReq(req *Request) string {
	_ = req
	// Harness default date used by monotree fixtures (YYYY-MM-DD).
	return "2026-06-30"
}

func Run(t *testing.T, req *Request) (*Response, error) {
	args := append([]string(nil), req.Args...)
	// ExtraEnv (e.g. file:// GOPROXY for apply tidy) requires process isolation → L3 binary.
	if len(req.ExtraEnv) > 0 {
		return runCLIWithEnv(t, req.RepoDir, req.WrkHome, args, propTagsWrkEnv(req))
	}
	// L2 in-process: WrkHome/Dir only (no ExtraEnv/Setenv/Chdir).
	var stdout, stderr bytes.Buffer
	code := wrkcli.RunCLI(args, wrkcli.RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Dir:     req.RepoDir,
		WrkHome: req.WrkHome,
		WrkDate: wrkDateForReq(req),
	})
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: code,
	}, nil
}


```
