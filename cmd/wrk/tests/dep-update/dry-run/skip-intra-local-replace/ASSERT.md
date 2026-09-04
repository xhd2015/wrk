## Expected Output

```
==== dep-update (dry-run) ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/app
      would: pin  example.com/dep  v0.0.1 -> v0.0.2
      would: go mod tidy(?:  \(local git(?:; go=go1\.\d+\.\d+; GOROOT=.+)?\)|  \(go=go1\.\d+\.\d+; GOROOT=.+\))?
  checkout  external/dep
    module  example.com/dep/cmd
      would: skip  example.com/dep  (intra-module replace)

dep-update: would update 1 modules, skipped 1 in 2 checkouts
```

## Expected

- Exit 0.
- Dry-run banner; `would: pin` + `would: go mod tidy` on primary.
- `would: skip` on cmd sub-module (intra-module replace); no `would: go mod tidy` for cmd.
- Summary `would update 1 modules, skipped 1 in 2 checkouts`.
- Both go.mods unchanged; no go.sum.

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

  checkout  \.
    module  example\.com/app
      would: pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      would: go mod tidy(?:  \(local git(?:; go=go1\.\d+\.\d+; GOROOT=.+)?\)|  \(go=go1\.\d+\.\d+; GOROOT=.+\))?
  checkout  external/dep
    module  example\.com/dep/cmd
      would: skip  example\.com/dep  \(intra-module replace\)

dep-update: would update 1 modules, skipped 1 in 2 checkouts
`)
	assertGoModUnchanged(t, req)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertNoTidyArtifacts(t, req)
}
```
