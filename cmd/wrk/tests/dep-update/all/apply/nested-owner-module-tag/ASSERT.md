## Expected

- Exit 0.
- Apply tree pin `example.com/lib/dep` v0.1.0 -> v0.2.0 (no argv `dep` header).
- Tidy ok for app; summary updated 1 in 1 checkouts.
- Require at clean version v0.2.0 (not the full tag path).

## Side Effects

- Nested tag prefix stripped to go require version.

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
	assertApplyBanner(t, resp.Stdout)
	assertNoArgvDepHeader(t, resp.Stdout)
	assertPinLine(t, resp.Stdout, modLibDep, req.WantOldVersion, req.WantVersion)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, wantCheckoutsOf(req), false)
	assertRequireVersion(t, req.ConsumerGoMod, modLibDep, req.WantVersion)
	body := readFile(t, req.ConsumerGoMod)
	assertNotContains(t, body, "packages/dep/v0.2.0")
	assertOwnerGoModUnchanged(t, req)
}
```
