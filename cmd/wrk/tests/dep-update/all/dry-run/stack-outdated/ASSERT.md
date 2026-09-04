## Expected Output

```
==== dep-update (dry-run) ====

  checkout  .
    module  example.com/app
      would: pin  example.com/lib  v1.0.0 -> v1.2.3
      would: go mod tidy(?:  \(local git(?:; go=go1\.22\.12; GOROOT=.+)?\)|  \(go=go1\.22\.12; GOROOT=.+\))?
  checkout  external/kool
    module  example.com/kool
      would: pin  example.com/lib  v1.0.0 -> v1.2.3
      would: go mod tidy(?:  \(local git(?:; go=go1\.22\.12; GOROOT=.+)?\)|  \(go=go1\.22\.12; GOROOT=.+\))?

dep-update: would update 2, already 0, skipped 0 in 2 checkouts
```

## Expected

- Exit 0.
- Dry-run tree; **no** argv `dep` header list.
- `would: pin` on the other stack checkout; zero writes.

## Side Effects

- Both go.mods unchanged; owner unchanged; no go.sum.

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
	assertNoArgvDepHeader(t, resp.Stdout)
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update \(dry-run\) ====

  checkout  \.
    module  example\.com/app
      would: pin  example\.com/lib  v1\.0\.0 -> v1\.2\.3
      would: go mod tidy(?:  \(local git(?:; go=go1\.\d+\.\d+; GOROOT=.+)?\)|  \(go=go1\.\d+\.\d+; GOROOT=.+\))?
  checkout  external/kool
    module  example\.com/kool
      would: pin  example\.com/lib  v1\.0\.0 -> v1\.2\.3
      would: go mod tidy(?:  \(local git(?:; go=go1\.\d+\.\d+; GOROOT=.+)?\)|  \(go=go1\.\d+\.\d+; GOROOT=.+\))?

dep-update: would update 2, already 0, skipped 0 in 2 checkouts
`)
	assertGoModUnchanged(t, req)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertOwnerGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
