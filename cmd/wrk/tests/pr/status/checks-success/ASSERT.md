## Expected Output

```
https://github.com/acme/app/pull/42
State:     open
Title:     Fix login
Checks:    success
Reviews:   review required
```

## Expected

- Exit code 0 (successful query).
- Stdout matches status block: URL, State=open, Title, Checks=success, Reviews=review required (exact labels; spacing via `prStatusStdout`).
- Stderr empty.
- Fake `gh`: `pr list` and `pr view` called; **`pr create` and `pr comment` not called**.
- No `pushed` / create / comment tokens on stdout.

## Side Effects

- Read-only: no git push; no PR create; no issue comment.

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

	want := prStatusStdout(prDefaultURL, "open", prExistingTitle, "success", "review required")
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))
	for _, tok := range []string{"PR created", "comment added", "title set", "body set", "pushed "} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("status must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "list")
	viewInv := assertGhSubcmdCalled(t, invocs, "view")
	assertGhArgContains(t, viewInv, fmt.Sprintf("%d", prExistingNumber))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
