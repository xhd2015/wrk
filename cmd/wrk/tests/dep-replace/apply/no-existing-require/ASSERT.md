## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/consumer
      replace  example.com/dep => <abs>

dep-replace: replaced in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Not-git nearest **D7**: absolute replace written despite no prior require.
- Apply tree as for single-dir.
- Baseline had no `require example.com/dep` (fixture guard).

## Side Effects

- go.mod gains replace only (D7); no tidy.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if strings.Contains(req.BaselineGoMod, "require "+modDep) {
		t.Fatalf("fixture bug: baseline should not require %s", modDep)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
__ABS__: type=string
---
==== dep-replace ====
dep  example\.com/dep => __ABS__

  checkout  \.
    module  example\.com/consumer
      replace  example\.com/dep => __ABS__

dep-replace: replaced in 1 modules in 1 checkouts
`)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertNoTidyArtifacts(t, req)
}
```
