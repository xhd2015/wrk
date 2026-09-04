## Expected Output

```
==== dep-update (dry-run) ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/consumer
      would: pin  example.com/dep  v0.0.1 -> v0.0.2
      would: go mod tidy(?:  \(local git(?:; go=go1\.22\.12; GOROOT=.+)?\)|  \(go=go1\.22\.12; GOROOT=.+\))?

dep-update: would update 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Dry-run banner + `would: pin` + `would: go mod tidy`.
- No bare `pin` apply lines.
- go.mod unchanged (replace still present); no go.sum.

## Side Effects

- None: dry-run must not mutate go.mod.

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
    module  example\.com/consumer
      would: pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      would: go mod tidy(?:  \(local git(?:; go=go1\.\d+\.\d+; GOROOT=.+)?\)|  \(go=go1\.\d+\.\d+; GOROOT=.+\))?

dep-update: would update 1 modules in 1 checkouts
`)
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
