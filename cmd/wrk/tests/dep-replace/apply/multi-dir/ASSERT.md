## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>
dep  example.com/dep2 => <abs2>

  checkout  .
    module  example.com/consumer
      replace  example.com/dep => <abs>
      replace  example.com/dep2 => <abs2>
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Two argv `dep` headers (argv order); one consumer; both replaces; versioned tidy after replaces.
- go.mod has absolute replaces for both modules.

## Side Effects

- Two absolute replaces written to consumer go.mod.

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
	assert.Output(t, resp.Stdout, `---
version: 3
__ABS__: type=string
__ABS2__: type=string
---
==== dep-replace ====
dep  example\.com/dep => __ABS__
dep  example\.com/dep2 => __ABS2__

  checkout  \.
    module  example\.com/consumer
      replace  example\.com/dep => __ABS__
      replace  example\.com/dep2 => __ABS2__
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
`)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep2, req.Dep2Dir)
}
```
