## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/consumer
      replace  example.com/dep => <abs>
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Apply banner + argv `dep` header; checkout `.`; `replace` line with absolute path.
- go.mod has absolute replace for `example.com/dep` (not `./` / `../`).
- No `would:` vocabulary.
- Includes `go mod tidy ok` after replaces (versioned tidy).

## Side Effects

- Consumer go.mod gains absolute replace to dep dir.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertNotContains(t, resp.Stdout, "would:")
	assert.Output(t, resp.Stdout, `---
version: 3
__ABS__: type=string
---
==== dep-replace ====
dep  example\.com/dep => __ABS__

  checkout  \.
    module  example\.com/consumer
      replace  example\.com/dep => __ABS__
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
`)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
}
```
