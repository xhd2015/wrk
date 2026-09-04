## Expected Output

```
==== dep-update ====

  checkout  .
    module  example.com/app
      pin  example.com/lib  v1.0.0 -> v1.2.3
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1, already 0, skipped 0 in 1 checkouts
```

## Expected

- Exit 0.
- Apply banner; **no** argv `dep` header list.
- Pin `example.com/lib` v1.0.0 -> v1.2.3; tidy ok for `example.com/app`.
- Summary `dep-update: updated 1, already 0, skipped 0 in 1 checkouts`.
- No `would:` vocabulary.
- Trailing newline.

## Side Effects

- Consumer require bumped to v1.2.3; go.sum exists after tidy.
- Owner go.mod unchanged (blast radius: stack Path only).
- No git commit required.

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
	assertNoArgvDepHeader(t, resp.Stdout)
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update ====

  checkout  \.
    module  example\.com/app
      pin  example\.com/lib  v1\.0\.0 -> v1\.2\.3
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1, already 0, skipped 0 in 1 checkouts
`)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertOwnerGoModUnchanged(t, req)
}
```
