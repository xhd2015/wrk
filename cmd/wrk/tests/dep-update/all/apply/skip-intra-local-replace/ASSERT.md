## Expected

- Exit 0.
- Apply tree pins inventory `example.com/lib` only.
- Summary `updated 1, already 0, skipped 1 in 1 checkouts`.
- No pin for `example.com/mono/lib`.
- Local replace for mono/lib retained; mono/lib require still v0.0.1.
- Inventory require bumped; tidy ok for mono.

## Side Effects

- Same-checkout filesystem replace → skipped (not bumped).

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
	assertCheckoutLine(t, resp.Stdout, checkoutLabelOf(req))
	assertPinLine(t, resp.Stdout, modLib, req.WantOldVersion, req.WantVersion)
	assertNoPinFor(t, resp.Stdout, modMonoLib)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, wantCheckoutsOf(req), false)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertRequireVersion(t, req.ConsumerGoMod, modMonoLib, "v0.0.1")
	assertReplacePresentFor(t, req.ConsumerGoMod, modMonoLib)
	assertOwnerGoModUnchanged(t, req)
}
```
