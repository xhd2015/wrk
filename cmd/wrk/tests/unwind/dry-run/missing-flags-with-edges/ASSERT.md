## Expected Output

Hard error (stderr preferred) naming required pin flags; no successful multi-peel plan.

```
wrk: ... --tag-next ... --push ...
```

(Exact wording implementer-owned; substrings locked.)

## Expected

- Exit code non-zero.
- Combined stdout/stderr mentions `--tag-next` (or `tag-next`) and `--push` (or `push`).
- Prefer no multi-step `would: peel` success plan.
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
	assertMissingPinFlagsError(t, resp)
	assertNoSuccessfulPeelPlan(t, resp.Stdout)
	assertUnwindZeroMutations(t, req)
}
```
