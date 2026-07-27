# wrk --projects-dep-graph — cross-project module dependency graph (P2)

## Version
0.0.2

Decision tree for root-level exclusive mode `wrk --projects-dep-graph`: print
registered projects, all Go modules (including nested), and **cross-project**
require edges only (among wrk-registered projects). Implementation must call
existing `wrkcli.BuildInventory` / `Inventory.CrossEdges` (P1); this tree
asserts **CLI behavior only**.

**Classic TDD (RED):** the flag and command do not exist yet. Leaves must fail
(unknown flag / non-zero / wrong output) until implementer wires CLI + formatter.

# DSN (Domain Specific Notion)

- **wrk CLI** — top-level exclusive mode flag `--projects-dep-graph` (like bare
  `--projects`). Sole mode flag: mutually exclusive with `--projects`, `--list`,
  and other modes. No git cwd required; uses `WRK_HOME` + `projects.json` only.
- **WRK_HOME registry** — `{WRK_HOME}/projects.json` lists main-repo absolute
  paths (same schema as `storage.ListProjects`). Paths sorted lexicographically
  for display order. Missing disk paths are soft-skipped with a stderr warning.
- **Inventory pipeline (P1, already landed)** — CLI invokes
  `BuildInventory(wrkHome)` then formats `Projects` + `CrossEdges()` (not
  `IntraEdges`, not external/unknown owners).
- **Human graph on stdout** — for each included project (sorted path order):

  ```text
  project <basename>  (<abs-path>)
    module <module-path>  dir=<rel-dir>
      → <dep-module>@<version>  [<owner-basename>]
  ```

  - Two spaces before `(` on the project line; module lines indented two spaces;
    cross-edge lines indented four spaces under the **consumer** module.
  - Modules listed under their project; nested modules use `dir=sub` style
    (`.` for root). Cross edges only (`→ dep@ver [ownerBasename]`).
  - Blank line between project blocks; blank line before the footer.
  - Footer (always, including empty registry):
    `N project(s)  ·  M module(s)  ·  K cross-edge(s)`
    with English pluralization (1 → singular form; else plural). Double spaces
    around `·`. Last content line ends with `\n`.
- **stderr soft issues** — missing registry path:
  `warning: project path does not exist: <abs-path>\n`
  (prefix `warning:`; still exit 0; remaining projects printed).
- **Hard mutual exclusion** — combine with `--projects` or `--list` → exit ≠ 0;
  stderr mentions exclusive / `wrk:`; stdout empty.
- **Help** — root `wrk -h` / `wrk --help` lists:
  `--projects-dep-graph` with description containing module-level dep graph /
  registered projects wording.
- **events.jsonl** — successful `--projects-dep-graph` appends
  `command: "projects-dep-graph"`.
- **Colors** — TTY policy only; doctests use pipes → plain text, no ANSI.
- **Out of scope** — `--propagate-tags`, external-only deps as edges, JSON,
  dry-run, pipeline / tag-next.

## Tree Overview

```
projects-dep-graph/
├── help/
│   └── mentions-flag/              # wrk -h documents --projects-dep-graph
├── mutual-exclusion/
│   ├── with-projects/              # + --projects → exclusive error
│   └── with-list/                  # + --list → exclusive error
├── graph/                          # exclusive mode success paths
│   ├── empty-registry/             # no projects.json → footer zeros
│   ├── single-project-no-edges/    # one project, one module, no → lines
│   ├── multi-cross-nested/         # lib root+sub + app require lib → 1 cross-edge
│   └── missing-path-soft-warn/     # missing + good → warning stderr, exit 0
└── events/
    └── command-projects-dep-graph/ # events.jsonl command=projects-dep-graph
```

Split factor (MECE, significance-first):

1. **Invocation class** — help surface | mutual exclusion | graph success | events.
2. Within mutual-exclusion: conflicting peer flag (`--projects` vs `--list`).
3. Within graph: registry topology (empty | single no-edge | multi nested cross |
   soft-skip missing).

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| H1 | help/mentions-flag | `wrk -h` → exit 0; help mentions `--projects-dep-graph` |
| X1 | mutual-exclusion/with-projects | `wrk --projects-dep-graph --projects` → exit ≠0; exclusive |
| X2 | mutual-exclusion/with-list | `wrk --projects-dep-graph --list` → exit ≠0; exclusive |
| G1 | graph/empty-registry | empty registry → stdout footer only `0 projects · 0 modules · 0 cross-edges` |
| G2 | graph/single-project-no-edges | one registered module project → project+module lines; 0 cross-edges; no `→` |
| G3 | graph/multi-cross-nested | lib (root+sub) + app requires lib@v1.2.3 → nested modules + one `→ … [lib]` |
| G4 | graph/missing-path-soft-warn | missing path + good project → `warning:` on stderr; exit 0; good project on stdout |
| E1 | events/command-projects-dep-graph | success → events.jsonl `command: "projects-dep-graph"` |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/projects-dep-graph
doctest test -v ./cmd/wrk/tests/projects-dep-graph
doctest test ./cmd/wrk/tests/projects-dep-graph/graph/empty-registry
doctest test ./cmd/wrk/tests/projects-dep-graph/graph/multi-cross-nested
doctest test ./cmd/wrk/tests/projects-dep-graph/mutual-exclusion/with-projects
```

Expect **RED** until implementer lands `--projects-dep-graph` (Classic TDD).

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
	RepoDir  string // process cwd when running wrk (neutral non-git by default)
	Args     []string

	// Fixture paths filled by graph leaves for Assert templates.
	LibPath     string
	AppPath     string
	GoodPath    string
	MissingPath string
	SoloPath    string

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product binary.
	// Prefer for help / mutual-exclusion / early reject leaves that do not need a
	// process boundary. Leave false (default) for true L3 e2e integration.
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
			Env:  depGraphWrkEnv(req),
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
	cmd.Env = depGraphWrkEnv(req)

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