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
- Not rejected as mutually exclusive (`--push` + `--pr` compose).
- Stdout: full-push confirm then new-PR block (blank line between stages OK).
- Origin `refs/heads/feature-pr` equals linked worktree HEAD (tip updated — unlike bare `--pr`).
- Pre-run origin snapshot is **behind** final origin tip (proves full push, not ensure-skip).
- Fake `gh`: create + comment called.

## Side Effects

- Full branch push published the local-ahead commit.
- New PR created + comment added.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
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
		t.Fatalf("--push --pr still mutually exclusive; stderr=%q", resp.Stderr)
	}
	assertEmptyStderr(t, resp.Stderr)

	branch := req.WtBranch
	assert.Output(t, resp.Stdout, v2StdoutTemplate(prComposePushThenCreateStdout(branch, prDefaultTitle, prDefaultURL)))

	// Full push: origin tip must match local HEAD (and advance past pre-run snapshot).
	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)
	beforeBytes, readErr := os.ReadFile(filepath.Join(req.WorkRoot, "origin-feature-before"))
	if readErr != nil {
		t.Fatalf("read origin snapshot: %v", readErr)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := revParseRef(t, req.OriginBare, "refs/heads/"+branch)
	if after == before {
		t.Fatalf("origin/%s must advance under --push --pr when local was ahead; still %s", branch, after)
	}
	local := revParseHEAD(t, req.WtDir)
	if after != local {
		t.Fatalf("origin/%s %s != local HEAD %s after --push --pr", branch, after, local)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "create")
	_ = assertGhSubcmdCalled(t, invocs, "comment")
}
```
