## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2
dep  example.com/dep2 -> v0.1.1

  checkout  .
    module  example.com/consumer
      pin  example.com/dep  v0.0.1 -> v0.0.2
      pin  example.com/dep2  v0.1.0 -> v0.1.1
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Two argv `dep` headers (argv order); one consumer; both pins; tidy once.
- Both replaces dropped; requires at latest tags; go.sum exists.

## Side Effects

- Two modules unpinned from local replace to tagged requires.

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
    module  example\.com/consumer
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      pin  example\.com/dep2  v0\.1\.0 -> v0\.1\.1
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
`)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep2)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertRequireVersion(t, req.ConsumerGoMod, modDep2, req.WantVersion2)
	assertGoSumExists(t, req.ConsumerModDir)
	if strings.Count(resp.Stdout, "go mod tidy ok") != 1 {
		t.Fatalf("expected tidy once for one consumer; got:\n%s", resp.Stdout)
	}
}
```
