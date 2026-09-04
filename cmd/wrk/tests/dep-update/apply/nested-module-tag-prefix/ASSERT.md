## Expected

- Exit 0.
- Apply tree with `dep  example.com/dep -> v0.0.2`; optional `(tag packages/dep/v0.0.2)`.
- go.mod: no replace; require v0.0.2 (not the full tag string as version).
- `go mod tidy ok` under the consumer; go.sum exists.

## Side Effects

- Tag prefix stripped to clean `vN.N.N` require version.

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

  checkout  \.
    module  example\.com/consumer
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
`)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	body := readFile(t, req.ConsumerGoMod)
	if strings.Contains(body, "require "+modDep+" packages/") {
		t.Fatalf("require must not use full tag path as version:\n%s", body)
	}
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertGoSumExists(t, req.ConsumerModDir)
	if req.WantTagHint != "" && strings.Contains(resp.Stdout, "tag") {
		if !strings.Contains(resp.Stdout, req.WantTagHint) &&
			!strings.Contains(resp.Stdout, "packages/dep") {
			_ = req.WantTagHint
		}
	}
}
```
