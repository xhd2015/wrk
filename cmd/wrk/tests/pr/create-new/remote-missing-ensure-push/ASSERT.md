## Expected Output

```
pushed feature-pr → origin/feature-pr
PR created
title set: Fix login
comment added
https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Stdout matches ensure-push + new-PR shape (trailing newline).
- Stderr empty (non-TTY plain text).
- Origin `refs/heads/feature-pr` equals linked worktree HEAD.
- Fake `gh` called `pr create` and `pr comment` (not only list).

## Side Effects

- Remote head branch created via push.
- Open PR created via `gh pr create` with title; comment added via `gh pr comment`.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	branch := req.WtBranch
	assert.Output(t, resp.Stdout, v2StdoutTemplate(prCreatedWithPushStdout(branch, prDefaultTitle, prDefaultURL)))

	if req.OriginBare == "" || req.WtDir == "" {
		t.Fatal("OriginBare and WtDir must be set")
	}
	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)

	invocs := parseFakeGhLog(t, ghLogPath(req))
	createInv := assertGhSubcmdCalled(t, invocs, "create")
	assertGhArgContains(t, createInv, prDefaultTitle)
	_ = assertGhSubcmdCalled(t, invocs, "comment")
	_ = assertGhSubcmdCalled(t, invocs, "list")
}
```
