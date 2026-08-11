## Expected

- Exit 0.
- Pin line for inventory `example.com/lib` → `v1.2.3`.
- Summary `updated 1, already 0, skipped 1`.
- No pin line for `example.com/mono/lib`.
- Local replace for mono/lib retained; mono/lib require still v0.0.1.
- Inventory require bumped; tidy ok for mono.

## Side Effects

- Same-toplevel filesystem replace → skipped (not bumped).

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
	// Local module must not appear as a pin action.
	assertNotContains(t, resp.Stdout, "dep-update "+modMonoLib)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, false)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertRequireVersion(t, req.ConsumerGoMod, modMonoLib, "v0.0.1")
	assertReplacePresentFor(t, req.ConsumerGoMod, modMonoLib)
	assertOwnerGoModUnchanged(t, req)
}
```
