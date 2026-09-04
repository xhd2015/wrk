## Expected

- Exit 0.
- Apply tree: banner, `dep` header, checkout `.`, pin, `go mod tidy ok`.
- go.sum exists; require @ latest.
- Wrapper at `go1.19.13` ran (recorded GOROOT/PATH0 contains that pin).

## Side Effects

- Tidy invoked via withgo pin `go1.19.13`, not the host default SDK.

## Exit Code

- 0

```go
import (
	"path/filepath"
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

  checkout  \.
    module  example\.com/consumer
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
`)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertVersionedGoUsed(t, req)
	goRoot := filepath.Join(req.InstallDir, req.WantGoPin)
	assertContains(t, resp.Stderr, "GOROOT="+goRoot)
	assertContains(t, resp.Stderr, filepath.Join(goRoot, "bin", "go")+" -C "+req.ConsumerModDir+" mod tidy")
	if strings.Contains(resp.Stderr, "$ go -C "+req.ConsumerModDir+" mod tidy") {
		t.Fatalf("verbose output must identify the overridden SDK, got:\n%s", resp.Stderr)
	}
}
```
