## Expected Output

```
==== dep-replace (dry-run) ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/consumer
      would: replace  example.com/dep => <abs>
      would: go mod tidy

dep-replace: would replace in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Dry-run banner + `would: replace` with absolute path.
- No bare `replace` apply lines.
- go.mod unchanged; dry-run plans tidy (`would: go mod tidy`) without writing.

## Side Effects

- None: dry-run must not mutate go.mod.

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
    module  example\.com/consumer
      would: replace  example\.com/dep => __ABS__
      would: go mod tidy(?:  \(go=go1\.\d+\.\d+; GOROOT=.+\))?

dep-replace: would replace in 1 modules in 1 checkouts
`)
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
