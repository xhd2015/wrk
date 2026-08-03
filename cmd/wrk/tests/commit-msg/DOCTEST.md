# wrk --commit -m/--message — manual commit message (no AI)

## Version
0.0.2

Decision tree for **manual** commit messages via `wrk --commit -m <msg>` /
`wrk --commit --message <msg>`. This is the wrk-owned commit path (no agent-pro
`gen-commit-msg` AI). Message sources are XOR: AI (`--gen-commit-msg`) vs manual
(`-m`/`--message` with `--commit`).

AI generate/commit regressions stay under `cmd/wrk/tests/gen-commit-msg/`.
Primary compose flag matrix remains primarily under `cmd/wrk/tests/done-compose/`;
this tree pins **manual** message + compose allow/reject at the flag layer and
standalone apply/dry-run/validation.

# DSN (Domain Specific Notion)

- **wrk CLI** — session-built binary or `wrkcli.Capture` (L2). Manual path:
  `wrk --commit -m MSG` / `--message MSG` with optional `--no-verify`,
  `--add-all`, `--dry-run`.
- **Message sources (XOR)** — AI: `--gen-commit-msg [--commit] …`; manual:
  `--commit` + `-m`/`--message`. Combining gen + `-m` is rejected.
- **Bare --commit** — without gen or `-m`/`--message` → hard error (needs a
  message source).
- **-m requires --commit** — `-m`/`--message` alone (no `--commit`) → error.
- **Empty message** — empty or whitespace-only `-m` value → reject.
- **Apply** — staged changes + manual message → `git commit` / CommitWithRetry;
  HEAD subject = message first line; multi-line body allowed in one string.
- **--dry-run** — plans `would: git commit` (+ message); no HEAD mutation.
- **--add-all** — stages all then commits (manual path).
- **--no-verify** — skips hooks when committing with manual message.
- **Shared branch** — multi-checkout of same branch → refuse commit (same
  framing as gen-commit-msg commit refuse: `Error:` / shared / refuse).
- **Compose partners** — same as gen-commit-msg: `--done`, `--merge-back`,
  `--sync`, `--tag-next`, `--push`, `--pr`, `--unwind`, `--reinstall-local`,
  `--exec`. Flag-layer pins: `--commit -m … --done` allowed; exclusive reject
  for gen + `-m`.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`.
- **Git cwd** — apply leaves use isolated repo under WorkRoot; process Dir is
  the repo (`Capture` / child `cmd.Dir`).

## Tree Overview

```
commit-msg/
├── help/                                    # wrk -h documents -m/--message
├── validation/
│   ├── m-requires-commit/                   # -m alone → requires --commit
│   ├── message-requires-commit/             # --message alone → requires --commit
│   ├── exclusive-with-gen-commit-msg/       # gen + -m → mutually exclusive
│   ├── commit-needs-message-source/         # --commit alone → need -m or gen
│   ├── empty-message/                       # --commit -m "" → empty message
│   └── whitespace-message/                  # --commit -m "   " → invalid
├── apply/
│   ├── succeeds/                            # staged + -m "feat: x" → HEAD subject
│   ├── succeeds-long-message/               # --message long form → HEAD subject
│   ├── multiline-message/                   # newline in -m string → subject line
│   ├── dry-run-no-mutate/                   # --dry-run → would: git commit; HEAD same
│   ├── no-verify/                           # failing pre-commit + --no-verify
│   ├── no-staged/                           # clean tree → nothing to commit
│   └── add-all/                             # untracked + --add-all → commits
├── shared-branch/
│   └── refuse/                              # multi-wt same branch → refuse
└── compose/
    ├── allow-done/                          # --commit -m --done not mutex
    └── allow-done-dry-run/                  # + --dry-run flag-layer accept
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| H1 | help | `wrk -h` documents `-m`/`--message`, requires `--commit`, exclusive with gen |
| V1 | validation/m-requires-commit | `-m "x"` → non-zero; requires `--commit` |
| V2 | validation/message-requires-commit | `--message "x"` → non-zero; requires `--commit` |
| V3 | validation/exclusive-with-gen-commit-msg | `--gen-commit-msg --commit --message "x"` → exclusive |
| V4 | validation/commit-needs-message-source | `--commit` alone → need message source |
| V5 | validation/empty-message | `--commit -m ""` → empty/invalid message |
| V6 | validation/whitespace-message | `--commit -m "   "` → empty/invalid message |
| S1 | apply/succeeds | staged + `--commit -m "feat: x"` → HEAD subject `feat: x` |
| S2 | apply/succeeds-long-message | staged + `--commit --message "feat: long form"` → HEAD |
| S3 | apply/multiline-message | `-m "feat: subj\n\nbody"` → HEAD subject `feat: subj` |
| S4 | apply/dry-run-no-mutate | `--dry-run` → would: git commit; HEAD unchanged |
| S5 | apply/no-verify | failing pre-commit + `--no-verify` → commit succeeds |
| S6 | apply/no-staged | clean + `--commit -m "x"` → non-zero; nothing to commit |
| S7 | apply/add-all | untracked + `--add-all` → stages and commits |
| S8 | shared-branch/refuse | multi-wt + `--commit -m "x"` → refuse |
| C1 | compose/allow-done | `--commit -m "x" --done` not mutually exclusive |
| C2 | compose/allow-done-dry-run | `--commit -m "x" --done --dry-run` flag-layer accept |

## How to Run

```sh
cd /Users/xhd2015/.wrk/worktrees/wrk-master-2026-08-03-add-m-msg-flag-which-requires-commit-and-is-exclusive-with-gen-c
doctest vet ./cmd/wrk/tests/commit-msg
doctest test -v ./cmd/wrk/tests/commit-msg
doctest test -v ./cmd/wrk/tests/commit-msg/help
doctest test -v ./cmd/wrk/tests/commit-msg/validation
doctest test -v ./cmd/wrk/tests/commit-msg/apply/succeeds
doctest test -v ./cmd/wrk/tests/commit-msg/compose/allow-done
```

Classic TDD: expect **RED** until implementer lands `--commit -m/--message`
validation, apply, dry-run, shared-branch refuse, and help text.

```go
import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot    string
	WrkHome     string
	RepoDir     string // process cwd when running wrk
	Args        []string
	HEADSubject string // pre-run HEAD subject for dry-run / refuse leaves
	MainRepo    string // shared-branch / compose fixtures
	WtDir       string
	Wt2Dir      string
	WtBranch    string
	ExtraEnv    []string

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product binary.
	// Prefer for validation / help / compose flag-layer / apply when Capture-safe.
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
			Env:  commitMsgWrkEnv(req),
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
	cmd.Env = commitMsgWrkEnv(req)

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
