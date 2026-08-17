## Expected Output

```
dep-update example.com/lib -> v1.2.3
skip tidy  module example.com/app  (vendor/)
dep-update: updated 1, already 0, skipped 0
```

## Expected

- Exit 0.
- Pin line for `example.com/lib` → `v1.2.3`.
- `skip tidy  module example.com/app  (vendor/)`.
- Summary `dep-update: updated 1, already 0, skipped 0` (pin selection unchanged).
- No `go mod tidy ok`; no `go mod vendor`.
- Require bumped; no go.sum; vendor/ has no `modules.txt`.
- Owner go.mod unchanged.

## Side Effects

- Same `--all` pin selection; tidy helper skips when vendor/ is present.

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
	assertNotContains(t, resp.Stdout, "would:")
	assertDepUpdateLine(t, resp.Stdout, modLib, req.WantVersion)
	assertSkipTidyLine(t, resp.Stdout, req.WantConsumerModule)
	assertNotContains(t, resp.Stdout, "go mod tidy ok")
	assertNotContains(t, resp.Stdout, "go mod vendor")
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, false)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertNoTidyArtifacts(t, req)
	assertVendorUntouched(t, req.VendorDir)
	assertOwnerGoModUnchanged(t, req)
}
```
