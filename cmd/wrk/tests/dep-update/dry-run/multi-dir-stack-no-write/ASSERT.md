## Expected Output

```
==== dep-update (dry-run) ====
dep  example.com/dep -> v0.0.2
dep  example.com/dep2 -> v0.1.1

  checkout  .
    module  example.com/app
      would: pin  example.com/dep  v0.0.1 -> v0.0.2
      would: pin  example.com/dep2  v0.1.0 -> v0.1.1
      would: go mod tidy(?:  \(go=go1\.22\.12; GOROOT=.+\))?
  checkout  external/kool
    module  example.com/kool
      would: pin  example.com/dep  v0.0.1 -> v0.0.2
      would: go mod tidy(?:  \(go=go1\.22\.12; GOROOT=.+\))?

dep-update: would update 2 modules in 2 checkouts
```

## Expected

- Exit 0.
- Two `dep` headers; would: pins; kool has only the first dep.
- Both go.mods unchanged.

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
---
==== dep-update \(dry-run\) ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?
dep  example\.com/dep2 -> v0\.1\.1(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/app
      would: pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      would: pin  example\.com/dep2  v0\.1\.0 -> v0\.1\.1
      would: go mod tidy(?:  \(go=go1\.\d+\.\d+; GOROOT=.+\))?
  checkout  external/kool
    module  example\.com/kool
      would: pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      would: go mod tidy(?:  \(go=go1\.\d+\.\d+; GOROOT=.+\))?

dep-update: would update 2 modules in 2 checkouts
`)
	assertGoModUnchanged(t, req)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertNoTidyArtifacts(t, req)
}
```
