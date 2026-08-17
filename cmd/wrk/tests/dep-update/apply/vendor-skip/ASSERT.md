## Expected Output

```
dep-update example.com/dep -> v0.0.2
skip tidy  module example.com/consumer  (vendor/)
```

## Expected

- Exit 0.
- Pin applied (replace dropped; require @ latest).
- Stdout `skip tidy  module example.com/consumer  (vendor/)`.
- No `go mod tidy ok`; no `go mod vendor`.
- No go.sum; vendor/ has no `modules.txt`.

## Side Effects

- Pin only; tidy and `go mod vendor` skipped.

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
	assertDepUpdateLine(t, resp.Stdout, modDep, req.WantVersion)
	assertSkipTidyLine(t, resp.Stdout, req.WantConsumerModule)
	assertNotContains(t, resp.Stdout, "go mod tidy ok")
	assertNotContains(t, resp.Stdout, "go mod vendor")
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertNoTidyArtifacts(t, req)
	assertVendorUntouched(t, req.VendorDir)
}
```
