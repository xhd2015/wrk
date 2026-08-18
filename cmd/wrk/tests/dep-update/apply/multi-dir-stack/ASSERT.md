## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2
dep  example.com/dep2 -> v0.1.1

  checkout  .
    module  example.com/app
      pin  example.com/dep  v0.0.1 -> v0.0.2
      pin  example.com/dep2  v0.1.0 -> v0.1.1
      go mod tidy ok
  checkout  external/kool
    module  example.com/kool
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok

dep-update: updated 2 modules in 2 checkouts
```

## Expected

- Exit 0.
- Two `dep` headers in argv order.
- Primary lists both pins then one tidy; kool lists only the first dep then one tidy.
- kool go.mod has no dep2 require.

## Side Effects

- Two modules updated across two checkouts; tidy once per affected module.

## Exit Code

- 0

```go
import (
	"strings"

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
dep  example\.com/dep2 -> v0\.1\.1(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/app
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      pin  example\.com/dep2  v0\.1\.0 -> v0\.1\.1
      go mod tidy ok
  checkout  external/kool
    module  example\.com/kool
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok

dep-update: updated 2 modules in 2 checkouts
`)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertRequireVersion(t, req.ConsumerGoMod, modDep2, req.WantVersion2)
	assertRequireVersion(t, req.Consumer2GoMod, modDep, req.WantVersion)
	kool := readFile(t, req.Consumer2GoMod)
	if strings.Contains(kool, modDep2) {
		t.Fatalf("kool must not require dep2:\n%s", kool)
	}
	if strings.Count(resp.Stdout, "go mod tidy ok") != 2 {
		t.Fatalf("expected tidy once per affected module; got:\n%s", resp.Stdout)
	}
	assertGoSumExists(t, req.ConsumerModDir)
	assertGoSumExists(t, req.Consumer2ModDir)
}
```
