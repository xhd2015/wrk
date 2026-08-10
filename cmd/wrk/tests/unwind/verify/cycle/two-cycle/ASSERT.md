## Expected Output

```
wrk: ... cycle ...
```

(Exact wording implementer-owned; must include substring `cycle`.)

## Expected

- Exit code non-zero.
- Stderr and/or stdout mentions `cycle`.
- No successful verify banners / `result:`.
- Zero mutations: host wt + both externals still present; HEADs match baseline.

## Side Effects

- None.

## Exit Code

- non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertCycleError(t, resp)
	assertNoSuccessfulVerifyBody(t, resp.Stdout)
	assertVerifyZeroMutations(t, req)
}
```
