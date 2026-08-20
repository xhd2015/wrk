## Expected Output

```
==== dep-update (dry-run) ====

  checkout  .
    module  example.com/app
      would: pin  example.com/lib  v1.0.0 -> v1.2.3
      would: go mod tidy(?:  \(go=go1\.22\.12; GOROOT=.+\))?

dep-update: would update 1, already 0, skipped 0 in 1 checkouts
```

## Expected

- Exit 0.
- Dry-run banner; **no** argv `dep` header list.
- `would: pin` for `example.com/lib` v1.0.0 -> v1.2.3.
- `would: go mod tidy` under the consumer module.
- Summary `dep-update: would update 1, already 0, skipped 0 in 1 checkouts`.
- Trailing newline on stdout.

## Side Effects

- Consumer go.mod unchanged; owner go.mod unchanged; no go.sum.

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
      would: go mod tidy

dep-update: would update 1, already 0, skipped 0 in 1 checkouts
`)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertGoModUnchanged(t, req)
	assertOwnerGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
