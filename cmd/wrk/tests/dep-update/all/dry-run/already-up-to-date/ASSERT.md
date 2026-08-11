## Expected Output

```
dep-update: already up to date
dep-update: updated 0, already 1, skipped 0
```

## Expected

- Exit 0.
- Banner `dep-update: already up to date`.
- Summary `dep-update: updated 0, already 1, skipped 0` (no-action form).
- No `would: dep-update` pin lines; no tidy would-lines required when no pins.
- Trailing newline.

## Side Effects

- Consumer and owner go.mod unchanged.

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
	assertAlreadyUpToDateBanner(t, resp.Stdout)
	// No-action summary uses apply wording per contract (updated 0, …).
	assertAllSummary(t, resp.Stdout, 0, req.WantAlready, req.WantSkipped, false)
	assertNotContains(t, resp.Stdout, "would: dep-update")
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertGoModUnchanged(t, req)
	assertOwnerGoModUnchanged(t, req)
}
```
