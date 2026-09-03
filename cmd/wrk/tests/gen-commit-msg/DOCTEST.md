# wrk --gen-commit-msg — CLI wire to agent-pro commit_msg

## Version
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
  `--commit`, `--no-verify`, `--agent-runner`, `--agent-runner-binary`, …).
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
├── help/                              # wrk --gen-commit-msg -h documents flags
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
│   └── with-sync/                     # bare --gen-commit-msg --sync → allowed compose
├── compose/
│   ├── clean-skip-with-exec/          # clean + --commit --exec → notice skip; exit 0
│   └── bare-clean-still-fails/        # clean + bare --commit → non-zero; no skip notice
└── validation/
    ├── no-verify-requires-commit/     # --no-verify without --commit → error
    └── unknown-agent-runner/          # --agent-runner codex → unsupported
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| H1 | help | `wrk --gen-commit-msg -h` → exit 0; usage tokens for mode + flags |
| D1 | dry-run/mock-message | staged 1 text file → mock B stdout; exit 0 |
| D2 | dry-run/accepts-model | `--dry-run --model some/model` → mock B success |
| D3 | dry-run/no-unstage-binary | binary+text staged → would-unstage on stderr; binary still staged; mock N=2 |
| D4 | dry-run/with-commit-no-mutate | `--dry-run --commit` → would: git commit; HEAD subject unchanged |
| D5 | dry-run/with-commit-no-verify-plan | `--dry-run --commit --no-verify` → would-line includes `--no-verify`; HEAD unchanged |
| G1 | generate/succeeds | fake-opencode mock → exit 0; stdout has `feat: add feature` + description |
| C1 | commit/succeeds | `--commit` + fake-opencode → HEAD subject `feat: add feature` |
| C2 | commit/no-verify | failing pre-commit + `--commit --no-verify` → subject `feat: skip hooks` |
| M1 | mutual-exclusion/with-status | `--gen-commit-msg --status` → non-zero; mutually exclusive |
| M2 | mutual-exclusion/with-sync | bare `--gen-commit-msg --sync` → allowed multi-stage compose |
| P1 | compose/clean-skip-with-exec | clean + `--add-all --gen-commit-msg --commit --exec true` → notice skip; exit 0 |
| P2 | compose/bare-clean-still-fails | clean + bare `--add-all --gen-commit-msg --commit` → non-zero; no skip notice |
| V1 | validation/no-verify-requires-commit | `--no-verify` alone → non-zero; requires --commit |
| V2 | validation/unknown-agent-runner | `--dry-run --agent-runner codex` → unsupported runner |

## How to Run

```sh
cd /Users/xhd2015/.wrk/worktrees/wrk-master-2026-07-16-wrk-gen-commit-msg-generate-commit-msg
doctest vet ./cmd/wrk/tests/gen-commit-msg
doctest test -v ./cmd/wrk/tests/gen-commit-msg
doctest test -v ./cmd/wrk/tests/gen-commit-msg/help
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
P2 compose allow/reject/help/dry-run leaves under monotree expect **RED** until implemented.

```go
import (
	"bytes"
	"os/exec"
	"testing"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot       string
	WrkHome        string
	RepoDir        string // process cwd when running wrk
	Args           []string
	BinaryRel      string // staged binary path (no-unstage-binary)
	HEADSubject    string // pre-run HEAD subject for dry-run --commit leaves
	ExtraEnv       []string // KEY=VAL for wrk (FAKE_OPENCODE_MOCK_CONFIG, OPENCODE_CONFIG_DIR)
	FakeOpencode   string // path to session-built fake-opencode
	MockConfigPath string // path written for FAKE_OPENCODE_MOCK_CONFIG

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
			Env:  genCommitMsgWrkEnv(req),
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
	cmd.Env = genCommitMsgWrkEnv(req)

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