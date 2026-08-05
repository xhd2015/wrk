# wrk skill — flag actions `--list` / `--show` / `--install`

## Version
0.0.3

Decision tree for `wrk skill`: early-dispatched flag actions that read the wrk
skill from `//go:embed SKILL.md` and optionally install `SKILL.md` to agent tool
directories. No git checkout is required.

**Layer:** L2 in-process CLI via `wrkcli.Capture` (short path — not binary e2e).

# DSN (Domain Specific Notion)

- **wrk CLI** — when `args[0] == "skill"`, intercept before git/worktree logic;
  `skill` is mutually exclusive with all other wrk modes/flags (`--done`,
  create, etc.). Skill-local flags (`--list`/`-l`, `--show`, `--install`) are
  skill actions, not wrk project-list modes.
- **Embedded SKILL.md** — `docs/skills/wrk/SKILL.md` is compiled via
  `//go:embed` (`docs/skills/wrk` package); no filesystem skills lookup at runtime.
- **Run** — `wrkcli.Capture` with `Dir=RepoDir` and `WRK_HOME` env only.
- **Shape 1 single skill** — no skill name argument; one embedded skill (`wrk`).
- **Action flags (exactly one)** — `--list` / `-l`, `--show`, or `--install`
  select the skill action. Subcommands `list` / `show` / `install` are rejected.
- **wrk skill --list | -l** — prints `wrk` (single line, trailing newline).
- **wrk skill --show [--header]** — prints embedded `SKILL.md` bytes, or YAML
  frontmatter only with `---` delimiters when `--header` is set; `--header` is
  only valid with `--show`; flag order of `--show` and `--header` may vary.
- **wrk skill --install [OPTIONS] [<dir>]** — installs embedded `SKILL.md` via
  `github.com/xhd2015/skills/install.HandleInstall` (`SkillDirName: "wrk"`,
  usage string `wrk skill --install`); `wrk skill --install --help` works.
- **Help / empty** — `wrk skill`, `wrk skill --help`, `wrk skill -h` print
  skill-level usage (mentions `--list`, `--show`, `--install`), exit 0;
  help alias spelling `-h,--help`; stdout ends with trailing `\n`.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; used for events only.

## Tree Overview

```
skill/
├── list/
│   ├── basic/                    # wrk skill --list → stdout wrk\n
│   └── alias-short/              # wrk skill -l → stdout wrk\n
├── show/
│   ├── basic/                    # wrk skill --show → embedded SKILL.md
│   ├── documents-new-flags/      # SKILL.md names polish flags + multi-mode --pr (not title always required)
│   ├── header/                   # --show --header → YAML block
│   ├── header-flag-order/        # --header --show → same as header
│   └── unknown-option/           # --show --nope → exit ≠0
├── install/
│   └── dry-run-cursor/           # --install --cursor --dry-run
├── help/
│   ├── empty/                    # wrk skill → usage, exit 0
│   ├── long/                     # wrk skill --help → usage, exit 0
│   └── short/                    # wrk skill -h → usage, exit 0
├── mutual-exclusion/
│   └── done/                     # skill --list --done → exit ≠0
└── reject-old-subcommand/
    └── list/                     # skill list → exit ≠0 (subcommand removed)
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| A1 | list/basic | `wrk skill --list` → exit 0; stdout `wrk\n` |
| A2 | list/alias-short | `wrk skill -l` → exit 0; stdout `wrk\n` |
| A3 | show/basic | `wrk skill --show` → exit 0; marker + `name: wrk`; trailing `\n` |
| A3b | show/documents-new-flags | `wrk skill --show` → polish flags + multi-mode `--pr` (status; title/comment not always-required) |
| A4 | show/header | `wrk skill --show --header` → YAML frontmatter; trailing `\n` |
| A5 | show/header-flag-order | `wrk skill --header --show` → same as A4 |
| A6 | install/dry-run-cursor | `wrk skill --install --cursor --dry-run` → dry-run, no writes |
| A7a | help/empty | `wrk skill` → exit 0; usage mentions `--list`/`--show`/`--install` |
| A7b | help/long | `wrk skill --help` → same skill-level usage |
| A7c | help/short | `wrk skill -h` → same skill-level usage |
| A8 | mutual-exclusion/done | `wrk skill --list --done` → exit ≠0; mutually exclusive |
| A9 | show/unknown-option | `wrk skill --show --nope` → exit ≠0; stderr unknown |
| A10 | reject-old-subcommand/list | `wrk skill list` → exit ≠0; clear rejection of subcommand |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/skill
doctest test ./cmd/wrk/tests/skill
doctest test ./cmd/wrk/tests/skill/list/basic
doctest test ./cmd/wrk/tests/skill/show/header
doctest test ./cmd/wrk/tests/skill/help/empty
doctest test ./cmd/wrk/tests/skill/reject-old-subcommand/list
doctest test ./cmd/wrk/tests/skill/install/dry-run-cursor
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // process cwd when running wrk
	Args     []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d // inject context available; Capture uses explicit Dir/Env only
	res := wrkcli.Capture(wrkcli.CaptureOpts{
		Args: append([]string(nil), req.Args...),
		Dir:  req.RepoDir,
		Env:  skillCaptureEnv(req),
	})
	return &Response{
		Stdout:   res.Stdout,
		Stderr:   res.Stderr,
		ExitCode: res.ExitCode,
	}, nil
}
```
