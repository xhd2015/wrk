## Expected

- Exit 0.
- Stdout has `dep-replace` lines for both `example.com/dep` and `example.com/dep2`.
- go.mod has absolute replaces for both modules.
- No tidy / no go.sum.

## Side Effects

- Two absolute replaces written to consumer go.mod.

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
	assertDepReplaceLine(t, resp.Stdout, modDep, req.DepDir)
	assertDepReplaceLine(t, resp.Stdout, modDep2, req.Dep2Dir)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep2, req.Dep2Dir)
	assertNoTidyArtifacts(t, req)
}
```
