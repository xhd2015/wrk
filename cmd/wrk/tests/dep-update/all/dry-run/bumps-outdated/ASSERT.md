## Expected Output

```
would: dep-update example.com/lib -> v1.2.3
would: go mod tidy  module example.com/app
dep-update: would update 1, already 0, skipped 0
```

## Expected

- Exit 0.
- Stdout has `would: dep-update` for `example.com/lib` → `v1.2.3`.
- Stdout has `would: go mod tidy` for consumer module `example.com/app`.
- Summary `dep-update: would update 1, already 0, skipped 0`.
- No bare apply `dep-update ` pin lines (without `would:`).
- Trailing newline on stdout.

## Side Effects

- Consumer go.mod unchanged; owner go.mod unchanged; no go.sum.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertWouldDepUpdateLine(t, resp.Stdout, modLib, req.WantVersion)
	assertWouldTidyLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, true)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertGoModUnchanged(t, req)
	assertOwnerGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
