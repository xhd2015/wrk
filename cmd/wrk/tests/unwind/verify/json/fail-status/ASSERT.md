## Expected

- Exit code non-zero (prefer 1).
- Stdout is pure JSON with shape keys and 6 checks.
- `require-drift` status is `fail`.
- `summary.result` is `fail`; `summary.fail` ≥ 1.
- No ANSI / human banners.
- No `Error:` for logical FAIL.
- Zero mutations.

## Side Effects

- None.

## Exit Code

- 1

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	_, checks, sum := assertVerifyJSONShape(t, resp.Stdout)
	assertVerifyJSONCheckStatus(t, checks, "require-drift", "fail")
	if sum.Result != "fail" {
		t.Fatalf("summary.result=%q want fail", sum.Result)
	}
	if sum.Fail < 1 {
		t.Fatalf("summary.fail=%d want >=1", sum.Fail)
	}
	assertVerifyNoLogicalErrorPrefix(t, resp)
	assertVerifyZeroMutations(t, req)
}
```
