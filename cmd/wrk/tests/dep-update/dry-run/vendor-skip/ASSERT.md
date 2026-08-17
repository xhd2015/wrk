## Expected Output

```
would: dep-update example.com/dep -> v0.0.2
would: skip tidy  module example.com/consumer  (vendor/)
```

## Expected

- Exit 0.
- `would: dep-update` and `would: skip tidy  module …  (vendor/)`.
- No `would: go mod tidy`; no bare apply lines.
- go.mod unchanged; no go.sum; vendor/ untouched.

## Side Effects

- None: dry-run must not mutate go.mod or vendor/.

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
	assertWouldDepUpdateLine(t, resp.Stdout, modDep, req.WantVersion)
	assertWouldSkipTidyLine(t, resp.Stdout, req.WantConsumerModule)
	assertNotContains(t, resp.Stdout, "would: go mod tidy")
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
	assertVendorUntouched(t, req.VendorDir)
}
```
