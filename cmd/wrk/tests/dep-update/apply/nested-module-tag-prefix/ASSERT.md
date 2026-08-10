## Expected

- Exit 0.
- Stdout contains `dep-update example.com/dep -> v0.0.2`.
- If tag parenthetical present, may include `packages/dep/v0.0.2` (implementer-owned).
- go.mod: no replace; require v0.0.2 (not the full tag string as version).

## Side Effects

- Tag prefix stripped to clean `vN.N.N` require version.

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
	assertDepUpdateLine(t, resp.Stdout, modDep, req.WantVersion)
	// Version token must be clean v0.0.2, not packages/dep/v0.0.2 as require version.
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	body := readFile(t, req.ConsumerGoMod)
	if strings.Contains(body, "require "+modDep+" packages/") {
		t.Fatalf("require must not use full tag path as version:\n%s", body)
	}
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	// Soft: if product prints tag form, prefer WantTagHint.
	if req.WantTagHint != "" && strings.Contains(resp.Stdout, "tag") {
		if !strings.Contains(resp.Stdout, req.WantTagHint) &&
			!strings.Contains(resp.Stdout, "packages/dep") {
			// Tag form is implementer-owned; do not hard-fail if only version printed.
			_ = req.WantTagHint
		}
	}
	assertNoTidyArtifacts(t, req)
}
```
