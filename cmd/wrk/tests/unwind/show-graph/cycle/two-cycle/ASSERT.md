## Expected Output

```
wrk: ... cycle ...
```

(Exact wording implementer-owned; must include substring `cycle`.)

## Expected

- Exit code non-zero.
- Stderr and/or stdout mentions `cycle`.
- No successful graph banners (`==== unwind graph (repo) ====`, module, summary).
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
	assertNoSuccessfulShowGraphBody(t, resp.Stdout)
	assertShowGraphZeroMutations(t, req)
}
```
