## Expected Output

```
Error: wrk: --show-graph cannot be used with --dry-run
```

(Exact wording implementer-owned; must mention show-graph and dry-run.)

## Expected

- Exit code non-zero.
- Combined stdout/stderr mentions `show-graph` and `dry-run`.
- No successful graph banners.
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
	assertShowGraphReject(t, resp, "dry-run")
	assertShowGraphZeroMutations(t, req)
}
```
