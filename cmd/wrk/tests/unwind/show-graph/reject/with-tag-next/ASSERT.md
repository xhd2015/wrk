## Expected Output

```
Error: wrk: --show-graph is mutually exclusive with --tag-next
```

(Exact wording implementer-owned; must mention show-graph and tag-next.)

## Expected

- Exit code non-zero.
- Combined output mentions `show-graph` and `tag-next`.
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
	assertShowGraphReject(t, resp, "tag-next")
	assertShowGraphZeroMutations(t, req)
}
```
