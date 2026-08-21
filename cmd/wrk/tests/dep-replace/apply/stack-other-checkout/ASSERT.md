## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/app
      replace  example.com/dep => <abs>
      go mod tidy ok
  checkout  external/kool
    module  example.com/kool
      replace  example.com/dep => <abs>
      go mod tidy ok

dep-replace: replaced in 2 modules in 2 checkouts
```

## Expected

- Exit 0.
- Absolute replace on primary **and** the other stack checkout that already requires dep.
- Checkout labels `.` and `external/kool`.
- kool filesystem replace on primary retained. Versioned tidy after replaces.

## Side Effects

- Both go.mods gain absolute replace to dep. No go.sum.

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
    module  example\.com/app
      replace  example\.com/dep => __ABS__
      go mod tidy ok
  checkout  external/kool
    module  example\.com/kool
      replace  example\.com/dep => __ABS__
      go mod tidy ok

dep-replace: replaced in 2 modules in 2 checkouts
`)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertAbsoluteReplace(t, req.Consumer2GoMod, modDep, req.DepDir)
}
```
