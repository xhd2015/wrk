## Expected Output

```
wrk: ... cycle ...
```

(Exact wording implementer-owned; must include substring `cycle`.)

## Expected

- Exit code non-zero.
- Stderr and/or stdout mentions `cycle`.
- No multi-step successful peel plan.
- No successful cascade body (`would: tag-next` absent on reject path).
- Zero mutations.

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
	assertNoSuccessfulCascadeBody(t, resp.Stdout)
	assertNoCascadeModuleLines(t, resp.Stdout)
	assertUnwindZeroMutations(t, req)
}
```
