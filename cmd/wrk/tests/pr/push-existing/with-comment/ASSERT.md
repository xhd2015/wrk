## Expected Output

```
pushed feature-pr → origin/feature-pr

comment added
https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Stdout: full-push confirm then comment-added + URL (blank line between stages OK).
- **No** `PR created` / `title set` / `body set`.
- Stderr empty — no title-ignored warning.
- Origin tip equals local HEAD and advanced past pre-run snapshot.
- Fake `gh`: `pr list` + `pr comment` (body = comment, PR number); **`pr create` not called**.

## Side Effects

- Full tip push then additive issue comment on existing open PR; never creates.

## Exit Code

- 0

```go
import (
	"fmt"
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
	assert.Output(t, resp.Stdout, v2StdoutTemplate(prPushExistingCommentStdout(branch, prDefaultURL)))
	for _, tok := range []string{"PR created", "title set", "body set"} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("push+comment must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}
	if strings.Contains(resp.Stderr, "title ignored") {
		t.Fatalf("push+comment must not warn title ignored; stderr=%q", resp.Stderr)
	}

	assertOriginBranchEqualsLocal(t, req.WtDir, req.OriginBare, branch)
	beforeBytes, readErr := os.ReadFile(filepath.Join(req.WorkRoot, "origin-feature-before"))
	if readErr != nil {
		t.Fatalf("read origin snapshot: %v", readErr)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := revParseRef(t, req.OriginBare, "refs/heads/"+branch)
	if after == before {
		t.Fatalf("origin/%s must advance under --push --pr --comment when local was ahead; still %s",
			branch, after)
	}
	local := revParseHEAD(t, req.WtDir)
	if after != local {
		t.Fatalf("origin/%s %s != local HEAD %s after push+comment", branch, after, local)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "list")
	commentInv := assertGhSubcmdCalled(t, invocs, "comment")
	assertGhArgContains(t, commentInv, prDefaultComment)
	assertGhArgContains(t, commentInv, fmt.Sprintf("%d", prExistingNumber))
	assertGhSubcmdNotCalled(t, invocs, "create")
}
```
