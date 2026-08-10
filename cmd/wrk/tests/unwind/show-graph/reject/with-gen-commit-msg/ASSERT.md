## Expected Output

```
Error: wrk: --show-graph is mutually exclusive with --gen-commit-msg
```

(Exact wording implementer-owned; must mention show-graph and gen-commit-msg.)

## Expected

- Exit code non-zero.
- Combined output mentions `show-graph` and `gen-commit-msg` (or gen-commit).
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
	// Accept gen-commit-msg or gen-commit as partner token.
	assertShowGraphReject(t, resp, "gen-commit")
	assertShowGraphZeroMutations(t, req)
}
```
