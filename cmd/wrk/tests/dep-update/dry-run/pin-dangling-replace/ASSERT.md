## Expected Output

```
==== dep-update (dry-run) ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/app
      would: pin  example.com/dep  v0.0.1 -> v0.0.2
      would: go mod tidy

dep-update: would update 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Dry-run banner; `would: pin` + `would: go mod tidy` on primary.
- No `would: skip ... (intra-module replace)` — the dangling replace target is
  not a git work tree, so it is never same-toplevel intra-module.
- Summary `would update 1 modules in 1 checkouts`.
- go.mod unchanged; no go.sum.

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
	assertNotContains(t, resp.Stdout, "(intra-module replace)")
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update \(dry-run\) ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/app
      would: pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      would: go mod tidy(?:  \(go=go1\.\d+\.\d+; GOROOT=.+\))?

dep-update: would update 1 modules in 1 checkouts
`)
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
