## Expected

- Exit 0.
- Apply tree pins only `example.com/lib` v1.0.0 -> v1.2.3.
- No pin for `example.com/external`.
- Summary `updated 1, already 0, skipped 0 in 1 checkouts` (external silent).
- Tidy ok for app; lib require bumped.

## Side Effects

- External require may remain in go.mod (silent skip); inventory dep updated.

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
	assertPinLine(t, resp.Stdout, modLib, req.WantOldVersion, req.WantVersion)
	assertNoPinFor(t, resp.Stdout, modExternal)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, wantCheckoutsOf(req), false)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertOwnerGoModUnchanged(t, req)
}
```
