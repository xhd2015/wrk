## Expected Output

```
pushed feature-pr → origin/feature-pr

https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Stdout: full-push confirm then URL only (blank line between stages OK via `joinStdoutBlocks`).
- **No** `PR created` / `title set` / `body set` / `comment added`.
- Stderr empty — no title-ignored warning (no title was passed).
- Origin `refs/heads/feature-pr` equals linked worktree HEAD (tip updated — full push).
- Pre-run origin snapshot is **behind** final origin tip (proves full push, not ensure-skip).
- Fake `gh`: `pr list` called; **`pr create` and `pr comment` not called**.

## Side Effects

- Full branch push published the local-ahead commit.
- Never creates a PR; never comments.

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
	assertEmptyStderr(t, resp.Stderr)

	branch := req.WtBranch
	assert.Output(t, resp.Stdout, v2StdoutTemplate(prPushExistingStdout(branch, prDefaultURL)))
	for _, tok := range []string{"PR created", "title set", "body set", "comment added"} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("push-existing bare must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}
	if strings.Contains(resp.Stderr, "title ignored") {
		t.Fatalf("push-existing must not warn title ignored; stderr=%q", resp.Stderr)
	}

	// Full push: origin tip must match local HEAD (and advance past pre-run snapshot).
	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)
	beforeBytes, readErr := os.ReadFile(filepath.Join(req.WorkRoot, "origin-feature-before"))
	if readErr != nil {
		t.Fatalf("read origin snapshot: %v", readErr)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := revParseRef(t, req.OriginBare, "refs/heads/"+branch)
	if after == before {
		t.Fatalf("origin/%s must advance under --pr --push when local was ahead; still %s", branch, after)
	}
	local := revParseHEAD(t, req.WtDir)
	if after != local {
		t.Fatalf("origin/%s %s != local HEAD %s after --pr --push", branch, after, local)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "list")
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
