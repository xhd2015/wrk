# wrk --gen-commit-msg — CLI wire to agent-pro commit_msg

## Version

**Layer: L2 in-process CLI** via `wrkcli.RunCLI`.
0.0.4

Decision tree for `wrk --gen-commit-msg`: top-level wrk mode that forwards to
`github.com/xhd2015/agent-pro/agent/commit_msg.RunGenCommitMsg`. Coverage backfill
for offline dry-run plan paths and non–dry-run generate/commit via fake-opencode.

**P2 primary compose** (pre-stage before `--done` / `--merge-back`) lives primarily in
monotree flag matrix `cmd/wrk/tests/done-compose/` and pipeline dry-run
`cmd/wrk/tests/done-pipeline/dry-run/with-gen-commit-msg/`. This tree keeps bare-mode
mutex pins (`--status`, bare `--sync`) and standalone generate/commit/dry-run.

# DSN (Domain Specific Notion)

- **wrk CLI** — session-built binary; bare `--gen-commit-msg` is a standalone mode
  mutually exclusive with other wrk modes (`--status`, bare `--sync`, `--list`, create, etc.).
  **P2**: with `--commit` it may compose as a **pre-stage** before `--done` / `--merge-back`
  (coverage under `done-compose/` + `done-pipeline/dry-run/with-gen-commit-msg/`).
- **Library path** — wrk imports only `agent/commit_msg` and calls
  `RunGenCommitMsg` with remaining gen-commit-msg flags (`--model`, `--dry-run`,
  `--commit`, `--no-verify`, `--add-all`, `--agent-runner`, `--agent-runner-binary`, …).
- **--add-all** — library bool flag: stage all changes (`git add -A`) before generate.
  Dry-run prints `would: git add -A` on stderr (no index mutation). Bare wrk path
  must forward the flag (not reject as unrecognized). Compose peel should treat it
  like `--commit` / `--no-verify` (`genCommitMsgBoolFlags`). Not wrk project `--add`.
- **--dry-run allow-list** — bare `wrk --dry-run` stays invalid; pair
  `--gen-commit-msg --dry-run` is valid (pure plan via library mock message B).
- **Mock message B** — stdout exact:
  `dry-run: would generate commit message for N staged file(s)\n`
  (N = staged file count **before** unstage; no agent, no index/HEAD mutation).
- **Dry-run would-unstage** — binaries/submodules planned on stderr
  (`would: unstage <path>…`); index is **not** mutated (binary stays staged).
- **Dry-run --commit** — stderr `would: git commit -m '…'` (and
  `--no-verify` when set); HEAD subject unchanged; never `Running git commit...`.
- **fake-opencode path** — session-built from
  `external/agent-pro-master-2026-07-16/cmd/fake-opencode` into
  `{fixtureSession}/bin/fake-opencode`. Leaves pass
  `--agent-runner opencode --agent-runner-binary <fake> --model openai/gpt-5`.
- **FAKE_OPENCODE_MOCK_CONFIG** — path to mock JSON
  (`agent-pro.fake-runner.v1` + `llm_events` with JSON title/description text);
  set on wrk process via `Request.ExtraEnv` so child fake-opencode inherits it.
- **OPENCODE_CONFIG_DIR** — temp dir under WorkRoot; set in ExtraEnv for
  hermetic fake-opencode session/config home.
- **Flag validation (library)** — `--no-verify` requires `--commit`;
  unknown `--agent-runner` (e.g. `codex`) → unsupported agent runner error
  even under `--dry-run`.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`.
- **Git cwd** — dry-run / generate success leaves init an isolated repo under
  WorkRoot with hooks disabled (`core.hooksPath=/dev/null`) unless a leaf needs
  real hooks (`commit/no-verify`); process cwd is the repo.

## Tree Overview

```
gen-commit-msg/
├── help/                              # wrk --gen-commit-msg -h documents flags (incl. --add-all)
├── add-all/
│   ├── dry-run-would-line/            # bare --add-all --dry-run → would: git add -A on stderr
│   └── compose-done-dry-run/          # RED: peel --add-all with --commit --done --dry-run
├── dry-run/
│   ├── mock-message/                  # staged 1 file → mock B; exit 0
│   ├── accepts-model/                 # --model some/model accepted under dry-run
│   ├── no-unstage-binary/             # binary+text staged → would-unstage; index unchanged; N=2
│   ├── with-commit-no-mutate/         # --dry-run --commit → would: git commit; HEAD unchanged
│   └── with-commit-no-verify-plan/    # --dry-run --commit --no-verify → would-line has --no-verify
├── generate/
│   └── succeeds/                      # fake-opencode; no --commit; stdout has title+description
├── commit/
│   ├── succeeds/                      # --commit; HEAD subject = mock title
│   └── no-verify/                     # failing pre-commit + --commit --no-verify → succeeds
├── mutual-exclusion/
│   ├── with-status/                   # --gen-commit-msg --status → mutex error
│   └── with-sync/                     # bare --gen-commit-msg --sync → mutex (no primary)
└── validation/
    ├── no-verify-requires-commit/     # --no-verify without --commit → error
    └── unknown-agent-runner/          # --agent-runner codex → unsupported
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| H1 | help | `wrk --gen-commit-msg -h` → exit 0; usage tokens for mode + flags incl. `--add-all` |
| A1 | add-all/dry-run-would-line | bare `--add-all --dry-run` → stderr `would: git add -A`; mock B; not unrecognized flag |
| A2 | add-all/compose-done-dry-run | compose `--add-all --commit --done --dry-run` peels flag; **RED** until `genCommitMsgBoolFlags` |
| D1 | dry-run/mock-message | staged 1 text file → mock B stdout; exit 0 |
| D2 | dry-run/accepts-model | `--dry-run --model some/model` → mock B success |
| D3 | dry-run/no-unstage-binary | binary+text staged → would-unstage on stderr; binary still staged; mock N=2 |
| D4 | dry-run/with-commit-no-mutate | `--dry-run --commit` → would: git commit; HEAD subject unchanged |
| D5 | dry-run/with-commit-no-verify-plan | `--dry-run --commit --no-verify` → would-line includes `--no-verify`; HEAD unchanged |
| G1 | generate/succeeds | fake-opencode mock → exit 0; stdout has `feat: add feature` + description |
| C1 | commit/succeeds | `--commit` + fake-opencode → HEAD subject `feat: add feature` |
| C2 | commit/no-verify | failing pre-commit + `--commit --no-verify` → subject `feat: skip hooks` |
| M1 | mutual-exclusion/with-status | `--gen-commit-msg --status` → non-zero; mutually exclusive |
| M2 | mutual-exclusion/with-sync | bare `--gen-commit-msg --sync` → non-zero; mutually exclusive (GREEN pin) |
| V1 | validation/no-verify-requires-commit | `--no-verify` alone → non-zero; requires --commit |
| V2 | validation/unknown-agent-runner | `--dry-run --agent-runner codex` → unsupported runner |

## How to Run

```sh
cd /Users/xhd2015/.wrk/worktrees/wrk-master-2026-07-16-wrk-gen-commit-msg-generate-commit-msg
doctest vet ./cmd/wrk/tests/gen-commit-msg
doctest test -v ./cmd/wrk/tests/gen-commit-msg
doctest test -v ./cmd/wrk/tests/gen-commit-msg/help
doctest test -v ./cmd/wrk/tests/gen-commit-msg/add-all/dry-run-would-line
doctest test -v ./cmd/wrk/tests/gen-commit-msg/add-all/compose-done-dry-run
doctest test -v ./cmd/wrk/tests/gen-commit-msg/dry-run/mock-message
doctest test -v ./cmd/wrk/tests/gen-commit-msg/generate/succeeds
doctest test -v ./cmd/wrk/tests/gen-commit-msg/commit/succeeds
doctest test -v ./cmd/wrk/tests/gen-commit-msg/commit/no-verify
doctest test -v ./cmd/wrk/tests/gen-commit-msg/mutual-exclusion/with-status
doctest test -v ./cmd/wrk/tests/gen-commit-msg/mutual-exclusion/with-sync
# P2 primary compose (flag + pipeline) — separate monotree roots:
doctest test -v ./cmd/wrk/tests/done-compose
doctest test -v ./cmd/wrk/tests/done-pipeline/dry-run/with-gen-commit-msg
```

Expect **GREEN** for standalone leaves (dry-run + fake-opencode generate/commit; no live LLM).
P2 bare `--add-all` help + dry-run-would-line pins expect **GREEN** when agent-pro already has
`--add-all` (P1) and bare wrk forwards library flags.
**A2 compose-done-dry-run** expects **RED** until implementer adds `--add-all` to
`genCommitMsgBoolFlags` (peel); root `wrk -h` may still need `--add-all` separately.

```go
import (
	"github.com/xhd2015/wrk/wrkcli"
	"strings"
	"bytes"
	"os/exec"
	"testing"
)

type Request struct {
	WorkRoot       string
	WrkHome        string
	RepoDir        string // process cwd when running wrk
	Args           []string
	BinaryRel      string // staged binary path (no-unstage-binary)
	HEADSubject    string // pre-run HEAD subject for dry-run --commit leaves
	MainRepo       string // main checkout for compose --done leaves
	WtBranch       string // linked worktree branch name for compose leaves
	ExtraEnv       []string // KEY=VAL for wrk (FAKE_OPENCODE_MOCK_CONFIG, OPENCODE_CONFIG_DIR)
	FakeOpencode   string // path to session-built fake-opencode
	MockConfigPath string // path written for FAKE_OPENCODE_MOCK_CONFIG
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
	// ExtraEnv (fake-opencode agent path) needs process isolation → L3 binary.
	if len(req.ExtraEnv) > 0 {
		return runCLIWithEnv(t, req.RepoDir, req.WrkHome, args, genCommitMsgWrkEnv(req))
	}
	// L2 in-process: WrkHome/Dir + writers (pipeline stages threaded via ctx).
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
