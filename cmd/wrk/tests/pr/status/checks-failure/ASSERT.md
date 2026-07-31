## Expected Output

```
https://github.com/acme/app/pull/42
State:     open
Title:     Fix login
Checks:    failure
Reviews:   review required
```

## Expected

- Exit code **0** even though Checks=failure (status is report-only, not a CI gate).
- Stdout Checks value is `failure`.
- Stderr empty.
- Fake `gh`: `pr list` + `pr view`; no create/comment.

## Side Effects

- Read-only; origin tip unchanged.

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
		t.Fatalf("status with failing checks must exit 0 (report not gate); got %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	want := prStatusStdout(prDefaultURL, "open", prExistingTitle, "failure", "review required")
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
