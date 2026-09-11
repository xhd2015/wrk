## Expected

- Exit code 0.
- Fully staged dirt: plan gen-commit + commit without `would: git add -A`.
- Zero mutations.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	out := resp.Stdout
	assertDryRunPlanShape(t, out)
	if strings.Contains(out, "would: git add -A") {
		t.Fatalf("without --add-all must not plan git add -A; stdout:\n%s", out)
	}
	assertContainsInOrder(t, out, "would: gen-commit-msg", "would: commit")
	assertUnwindZeroMutations(t, req)
}
```
