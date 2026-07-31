# Scenario

**Feature**: multi-stage compose partners with `--pr` — gen-commit → push → pr (fixed order)

```
# P2 partners (flag argv order free; execution order fixed)
[optional] --gen-commit-msg [--commit …] [--agent-runner …]
[optional] --push
           --pr --title T --comment C

# --push --pr: always full branch push first, then PR path
# --pr only (P1): ensure-push only when remote head missing — not this grouping
linked wt + github origin + fake gh
  -> wrk --push --pr --title T --comment C
  -> full git push of head branch (even if origin already has branch)
  -> then gh pr list/create/comment
  -> stdout: pushed … then PR tokens / URL

# gen-commit + push + pr (hermetic commandcode mock binary)
linked wt + staged file + commandcode shell mock
  -> wrk --gen-commit-msg --commit --agent-runner commandcode … --push --pr …
  -> stage 1: commit with mock title
  -> stage 2: full push
  -> stage 3: PR create/comment
```

## Preconditions

- Inherits `pr/SETUP.md` fake `gh`, github-shaped origin + bare pushurl fixtures.
- Classic TDD: compose is **not** implemented yet (`--pr` currently exclusive with `--push` / `--gen-commit-msg`). Leaves RED until implementer.
- L2 `InProcess = true`; parallel-safe (`PathPrepend` / `ExtraEnv` only).
- Reuse P1 fixtures: `setupPrLinkedFeature*`, `installFakeGh`, stdout/gh helpers.
- Gen-commit hermetic path: shell `--agent-runner-binary` with `--agent-runner commandcode` (no live LLM, no external fake-opencode tree).
- Real `--commit` leaves disable host hooks via repo-local `core.hooksPath=/dev/null` (same pattern as `gen-commit-msg/`).

## Steps

- Grouping only. Leaves set fixtures, optional staged file + commandcode mock, `req.Args`.

## Context

- Full `--push` uses the same confirm line as bare push / ensure-push: `pushed <branch> → origin/<branch>`.
- After full push, bare `--pr` ensure-push is a no-op (remote tip already matches).
- Multi-stage stdout may insert a blank line between stages (same style as `runActiveRootPipeline`); asserts accept blank separators via `joinStdoutBlocks`.

```go
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

const (
	// Mock commit title/description from commandcode shell agent.
	prComposeCommitTitle = "feat: compose pr"
	prComposeCommitDesc  = "staged for compose"
	prComposeCommitJSON  = `{"title":"feat: compose pr","description":"staged for compose"}`
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	prComposeEnsureHelpers()
	return nil
}

// prPushPrArgs builds --push --pr with default title/comment (push flag first).
func prPushPrArgs() []string {
	return append([]string{"--push"}, prDefaultArgs()...)
}

// prPrThenPushArgs: argv order free — --pr before --push (execution still push→pr).
func prPrThenPushArgs() []string {
	return append(prDefaultArgs(), "--push")
}

// prComposePushThenCreateStdout is the multi-stage shape: full-push confirm,
// blank line, then new-PR success block.
func prComposePushThenCreateStdout(branch, title, url string) string {
	return joinStdoutBlocks(prPushConfirmLine(branch), prCreatedStdout(title, url))
}

// installCommandCodeCommitMock writes a shell binary that prints fixed commit
// JSON for --agent-runner commandcode (no llm-mock build / no external agent-pro).
// Returns absolute path for --agent-runner-binary.
func installCommandCodeCommitMock(t *testing.T, req *Request) string {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "commandcode-bin")
	mkdirAll(t, binDir)
	bin := filepath.Join(binDir, "cmd-mock")
	// Ignore argv; always emit fixed JSON on stdout (commandcode full-stdout parse).
	body := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' %q\n", prComposeCommitJSON)
	if err := os.WriteFile(bin, []byte(body), 0o755); err != nil {
		t.Fatalf("write commandcode mock: %v", err)
	}
	return bin
}

// stagePrComposeFile stages one text file on the linked worktree for gen-commit.
func stagePrComposeFile(t *testing.T, req *Request) {
	t.Helper()
	if req.WtDir == "" {
		t.Fatal("stagePrComposeFile: WtDir empty")
	}
	path := filepath.Join(req.WtDir, "compose-stage.go")
	writeFile(t, path, "package compose\n// staged for gen-commit-msg compose\n")
	runGitIsolated(t, req.WtDir, "add", "compose-stage.go")
}

// disablePrRepoHooks sets repo-local core.hooksPath=/dev/null so real git commit
// (gen-commit-msg --commit) does not hit host global/author hooks.
// Prefer MainRepo so linked worktrees inherit the common git dir config.
func disablePrRepoHooks(t *testing.T, req *Request) {
	t.Helper()
	repo := req.MainRepo
	if repo == "" {
		repo = req.WtDir
	}
	if repo == "" {
		repo = req.RepoDir
	}
	if repo == "" {
		t.Fatal("disablePrRepoHooks: no MainRepo/WtDir/RepoDir")
	}
	runGitIsolated(t, repo, "config", "core.hooksPath", "/dev/null")
}

// prGenCommitPushPrArgs builds hermetic gen-commit + push + pr argv.
func prGenCommitPushPrArgs(agentBin string) []string {
	return []string{
		"--gen-commit-msg", "--commit",
		"--agent-runner", "commandcode",
		"--agent-runner-binary", agentBin,
		"--push",
		"--pr", "--title", prDefaultTitle, "--comment", prDefaultComment,
	}
}

func prComposeEnsureHelpers() {
	_ = prPushPrArgs
	_ = prPrThenPushArgs
	_ = prComposePushThenCreateStdout
	_ = installCommandCodeCommitMock
	_ = stagePrComposeFile
	_ = disablePrRepoHooks
	_ = prGenCommitPushPrArgs
}
```
