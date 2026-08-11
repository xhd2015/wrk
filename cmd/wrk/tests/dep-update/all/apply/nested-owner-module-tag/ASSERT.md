## Expected

- Exit 0.
- Pin line `dep-update example.com/lib/dep -> v0.2.0` (optional tag form may
  mention `packages/dep/v0.2.0`).
- Tidy ok for app; summary updated 1.
- Require at clean version v0.2.0 (not the full tag path).

## Side Effects

- Nested tag prefix stripped to go require version.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertNotContains(t, resp.Stdout, "would:")
	assertDepUpdateLine(t, resp.Stdout, modLibDep, req.WantVersion)
	if req.WantTagHint != "" && strings.Contains(resp.Stdout, "tag") {
		// Optional: when tag parenthetical present, prefer nested tag form.
		assertContains(t, resp.Stdout, req.WantTagHint)
	}
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, false)
	assertRequireVersion(t, req.ConsumerGoMod, modLibDep, req.WantVersion)
	// Require version must be clean vN.N.N, not packages/dep/vN.N.N.
	body := readFile(t, req.ConsumerGoMod)
	assertNotContains(t, body, "packages/dep/v0.2.0")
	assertOwnerGoModUnchanged(t, req)
}
```
