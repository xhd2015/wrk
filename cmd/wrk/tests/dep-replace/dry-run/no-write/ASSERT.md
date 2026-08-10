## Expected

- Exit 0.
- Stdout contains `would: dep-replace example.com/dep => ` with absolute path.
- No bare `dep-replace ` apply lines.
- go.mod unchanged; no go.sum created (no tidy).

## Side Effects

- None: dry-run must not mutate go.mod.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertWouldDepReplaceLine(t, resp.Stdout, modDep, req.DepDir)
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
