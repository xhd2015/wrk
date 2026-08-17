## Expected

- Exit 0.
- Stdout contains `would: dep-update example.com/dep -> v0.0.2`.
- Stdout contains `would: go mod tidy  module example.com/consumer`.
- No bare `dep-update ` apply lines.
- go.mod unchanged (replace still present); no go.sum.

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
	assertWouldDepUpdateLine(t, resp.Stdout, modDep, req.WantVersion)
	assertWouldTidyLine(t, resp.Stdout, req.WantConsumerModule)
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
