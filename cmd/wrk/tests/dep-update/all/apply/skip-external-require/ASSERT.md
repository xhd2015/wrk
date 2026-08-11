## Expected

- Exit 0.
- Pin only for `example.com/lib` → `v1.2.3`.
- No pin/mention action line for `example.com/external`.
- Summary `updated 1, already 0, skipped 0` (external silent).
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
	assertDepUpdateLine(t, resp.Stdout, modLib, req.WantVersion)
	assertNotContains(t, resp.Stdout, "dep-update "+modExternal)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, false)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertOwnerGoModUnchanged(t, req)
}
```
