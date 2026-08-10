## Expected

- Exit 0.
- Absolute replace written despite no prior require.
- Stdout has `dep-replace` line.
- Baseline had no `require example.com/dep` (fixture guard).

## Side Effects

- go.mod gains replace only (D7); no tidy.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	// Guard: baseline truly had no require for dep.
	if strings.Contains(req.BaselineGoMod, "require "+modDep) {
		t.Fatalf("fixture bug: baseline should not require %s", modDep)
	}
	assertDepReplaceLine(t, resp.Stdout, modDep, req.DepDir)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertNoTidyArtifacts(t, req)
}
```
