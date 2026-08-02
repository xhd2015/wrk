## Expected

- `err` is nil.
- `resp.Tag` is non-empty (planned next tag; expected `v0.0.2` for this seed).
- Tag ref `v0.0.2` does **not** exist on the main repo (DryRun only).

## Side Effects

- None (plan only; no lightweight tag created).

## Errors

- None.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if strings.TrimSpace(resp.Tag) == "" {
		t.Fatal("Tag empty after TagNext DryRun")
	}
	// Seed plans root patch bump from v0.0.1 → v0.0.2 (wrk/tagscope scheme).
	if resp.Tag != "v0.0.2" {
		t.Fatalf("Tag: got %q want %q", resp.Tag, "v0.0.2")
	}
	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 tag should not exist after DryRun")
	}
}
```
