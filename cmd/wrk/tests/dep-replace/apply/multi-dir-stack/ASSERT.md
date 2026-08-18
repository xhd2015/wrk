## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>
dep  example.com/dep2 => <abs2>

  checkout  .
    module  example.com/app
      replace  example.com/dep => <abs>
      replace  example.com/dep2 => <abs2>
  checkout  external/kool
    module  example.com/kool
      replace  example.com/dep => <abs>

dep-replace: replaced in 2 modules in 2 checkouts
```

## Expected

- Exit 0.
- Two `dep` headers in argv order.
- Primary lists both replaces; kool lists only the first dep.
- kool go.mod has no replace for dep2.

## Side Effects

- Two modules updated across two checkouts; no tidy.

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
    module  example\.com/app
      replace  example\.com/dep => __ABS__
      replace  example\.com/dep2 => __ABS2__
  checkout  external/kool
    module  example\.com/kool
      replace  example\.com/dep => __ABS__

dep-replace: replaced in 2 modules in 2 checkouts
`)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep2, req.Dep2Dir)
	assertAbsoluteReplace(t, req.Consumer2GoMod, modDep, req.DepDir)
	assertNoReplaceFor(t, req.Consumer2GoMod, modDep2)
	assertNoTidyArtifacts(t, req)
}
```
