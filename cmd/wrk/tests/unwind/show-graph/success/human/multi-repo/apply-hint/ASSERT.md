## Expected Output

```
==== status summary ====
…
hint: apply would need --merge-back --tag-next --push
```

(Exact hint wording implementer-owned; must mention the needed flags.)

## Expected

- Exit code **0** (missing apply flags is **not** an error on show-graph).
- Human banners present.
- Summary (or trailing section) hints that apply would need land + pin flags:
  `--merge-back` (or land), `--tag-next`, `--push`.
- Graph body printed (not a flag-validation hard fail).
- Zero mutations.

## Side Effects

- None.

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
	assertShowGraphHumanBanners(t, resp.Stdout)
	combined := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)
	// Read-only hint — not a hard Error path.
	if !strings.Contains(lower, "hint") && !strings.Contains(lower, "would need") &&
		!strings.Contains(lower, "apply would") {
		t.Fatalf("expected apply-needs hint language; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(lower, "tag-next") && !strings.Contains(combined, "--tag-next") {
		t.Fatalf("hint must mention --tag-next; out:\n%s", combined)
	}
	if !strings.Contains(lower, "push") && !strings.Contains(combined, "--push") {
		t.Fatalf("hint must mention --push; out:\n%s", combined)
	}
	// Land: merge-back or done language.
	if !strings.Contains(lower, "merge-back") && !strings.Contains(lower, "merge_back") &&
		!strings.Contains(combined, "--merge-back") && !strings.Contains(lower, "land") {
		t.Fatalf("hint must mention land/merge-back; out:\n%s", combined)
	}
	assertShowGraphZeroMutations(t, req)
}
```
