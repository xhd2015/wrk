## Expected Output

```
pushed feature-pr → origin/feature-pr

PR created
title set: Fix login
body set
https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Not mutually exclusive.
- Stdout: full-push confirm then new-PR block.
- Origin has `feature-pr` equal to linked HEAD.
- Fake `gh`: list + create (body = comment); no issue comment.

## Side Effects

- Remote head branch created via full push stage.
- New PR with body = `--comment`.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("--pr --push still mutually exclusive; stderr=%q", resp.Stderr)
	}
	assertEmptyStderr(t, resp.Stderr)

	branch := req.WtBranch
	assert.Output(t, resp.Stdout, v2StdoutTemplate(prComposePushThenCreateStdout(branch, prDefaultTitle, prDefaultURL)))

	if !originBranchExists(t, req.OriginBare, branch) {
		t.Fatalf("origin should have refs/heads/%s after --push --pr", branch)
	}
	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "list")
	createInv := assertGhSubcmdCalled(t, invocs, "create")
	assertGhArgContains(t, createInv, prDefaultTitle)
	assertGhArgContains(t, createInv, prDefaultComment)
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
