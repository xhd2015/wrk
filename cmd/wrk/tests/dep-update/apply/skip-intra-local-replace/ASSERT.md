## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/app
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok
  checkout  external/dep
    module  example.com/dep/cmd
      skip  example.com/dep  (intra-module replace)

dep-update: updated 1 modules, skipped 1 in 2 checkouts
```

## Expected

- Exit 0.
- Primary consumer pinned + tidied; replace to dep dropped.
- Cmd sub-module **skipped** (intra-module replace); no pin, no tidy.
- Cmd go.mod unchanged: require dep v0.0.1, replace dep => ../ retained.
- Summary `updated 1 modules, skipped 1 in 2 checkouts`.

## Side Effects

- Only the primary consumer is mutated; cmd sub-module untouched.

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
	assertApplyBanner(t, resp.Stdout)
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/app
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok
  checkout  external/dep
    module  example\.com/dep/cmd
      skip  example\.com/dep  \(intra-module replace\)

dep-update: updated 1 modules, skipped 1 in 2 checkouts
`)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertReplacePresentFor(t, req.Consumer2GoMod, modDep)
	assertRequireVersion(t, req.Consumer2GoMod, modDep, "v0.0.1")
}
```
