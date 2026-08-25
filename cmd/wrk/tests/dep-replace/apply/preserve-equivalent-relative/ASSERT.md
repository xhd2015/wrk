## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/app
      replace  example.com/dep => <abs>
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Primary gains absolute replace to dep.
- Nested `dep/cmd` go.mod unchanged (`replace => ../`).
- No module block for `example.com/dep` (self) or `example.com/dep/cmd` (equivalent skip).

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
	assert.Output(t, resp.Stdout, `---
version: 3
__ABS__: type=string
---
==== dep-replace ====
dep  example\.com/dep => __ABS__

  checkout  \.
    module  example\.com/app
      replace  example\.com/dep => __ABS__
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "module  "+modDep+"\n")
	assertNotContains(t, resp.Stdout, "module  example.com/dep/cmd")
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	body := readFile(t, req.Consumer2GoMod)
	if !strings.Contains(body, "replace "+modDep+" => ../") {
		t.Fatalf("nested cmd should keep relative replace; go.mod:\n%s", body)
	}
}
```
