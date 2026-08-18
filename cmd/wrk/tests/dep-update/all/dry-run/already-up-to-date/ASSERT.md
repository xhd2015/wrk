## Expected Output

```
==== dep-update ====
dep-update: already up to date
dep-update: updated 0, already 1, skipped 0 in 1 checkouts
```

## Expected

- Exit 0.
- Banner (`==== dep-update` prefix; dry-run banner also OK).
- `dep-update: already up to date`.
- Zero summary `dep-update: updated 0, already 1, skipped 0 in 1 checkouts`.
- No pin tree / no `would: pin`.
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
	assertAllSummary(t, resp.Stdout, 0, req.WantAlready, req.WantSkipped, wantCheckoutsOf(req), false)
	assertNotContains(t, resp.Stdout, "would: pin")
	assertNotContains(t, resp.Stdout, "pin  ")
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertGoModUnchanged(t, req)
	assertOwnerGoModUnchanged(t, req)
}
```
