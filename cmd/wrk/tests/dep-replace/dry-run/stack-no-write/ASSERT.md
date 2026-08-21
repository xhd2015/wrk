## Expected Output

```
==== dep-replace (dry-run) ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/app
      would: replace  example.com/dep => <abs>
      would: go mod tidy
  checkout  external/kool
    module  example.com/kool
      would: replace  example.com/dep => <abs>
      would: go mod tidy

dep-replace: would replace in 2 modules in 2 checkouts
```

## Expected

- Exit 0.
- Single-target dry-run stack tree; `would: replace` on both checkouts.
- Both go.mods unchanged; dry-run plans tidy without writing.

## Side Effects

- Zero writes.

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
---
==== dep-replace \(dry-run\) ====
dep  example\.com/dep => __ABS__

  checkout  \.
    module  example\.com/app
      would: replace  example\.com/dep => __ABS__
      would: go mod tidy(?:  \(go=go1\.\d+\.\d+; GOROOT=.+\))?
  checkout  external/kool
    module  example\.com/kool
      would: replace  example\.com/dep => __ABS__
      would: go mod tidy(?:  \(go=go1\.\d+\.\d+; GOROOT=.+\))?

dep-replace: would replace in 2 modules in 2 checkouts
`)
	assertGoModUnchanged(t, req)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertNoTidyArtifacts(t, req)
}
```
