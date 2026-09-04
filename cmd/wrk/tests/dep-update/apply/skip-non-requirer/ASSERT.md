## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/app
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Pin + tidy for `example.com/app` only (default quiet: no skip line).
- No `module  example.com/other`; no `no require` skip vocabulary.
- Sibling go.mod identical to baseline (no new require).

## Side Effects

- Only modules that already required xxx are mutated.

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
==== dep-update ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/app
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "module  "+req.WantConsumer2Module)
	assertNotContains(t, resp.Stdout, "no require")
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertGoSumExists(t, req.ConsumerModDir)
}
```
