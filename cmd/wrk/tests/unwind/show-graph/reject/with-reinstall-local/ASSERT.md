## Expected Output

```
Error: wrk: --show-graph is mutually exclusive with --reinstall-local
```

(Exact wording implementer-owned; must mention show-graph and reinstall-local.)

## Expected

- Exit code non-zero.
- Combined output mentions `show-graph` and `reinstall-local`.
- No successful graph banners; zero mutations.

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
	assertShowGraphReject(t, resp, "reinstall-local")
	assertShowGraphZeroMutations(t, req)
}
```
