## Expected Output

```
# stdout:
comment added
https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Stdout is comment-added + URL only (same tokens as existing-PR attach path; **no** `PR created` / `title set` / `pushed`).
- Stderr **empty** — no title-ignored warning (no title was passed).
- Fake `gh`: `pr list` and `pr comment` called with body = comment and PR number; **`pr create` not called**.
- No ensure-push / git push of tip.

## Side Effects

- Additive issue comment on existing open PR only; never creates a PR; never pushes.

## Exit Code

- 0

```go
import (
	"fmt"
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

	assert.Output(t, resp.Stdout, v2StdoutTemplate(prExistingStdout(prDefaultURL)))
	for _, tok := range []string{"PR created", "title set", "body set", "pushed "} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("comment-only must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}
	if strings.Contains(resp.Stderr, "title ignored") {
		t.Fatalf("comment-only must not warn title ignored; stderr=%q", resp.Stderr)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "list")
	commentInv := assertGhSubcmdCalled(t, invocs, "comment")
	assertGhArgContains(t, commentInv, prDefaultComment)
	assertGhArgContains(t, commentInv, fmt.Sprintf("%d", prExistingNumber))
	assertGhSubcmdNotCalled(t, invocs, "create")
}
```
