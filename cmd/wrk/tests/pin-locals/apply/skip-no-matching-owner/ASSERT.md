## Expected

- Exit 0.
- No replace for `example.com/missing` or `example.com/dep`.
- No pin-local success lines inventing edges.
- Already-up-to-date style or applied 0 summary.

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
	assertNoReplaceFor(t, req.ConsumerGoMod, modMissing)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertNotContains(t, resp.Stdout, "<- "+modMissing)
	// Should not invent pin of unrequired dep either.
	assertNotContains(t, resp.Stdout, "pin-local "+modConsumer+" <- "+modDep)
	// Prefer already message or applied 0.
	all := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if strings.Contains(resp.Stdout, "pin-locals:") {
		assertSummaryApplied(t, resp.Stdout, 0, 0, 0)
	} else if !strings.Contains(all, "already") && !strings.Contains(all, "up to date") {
		if strings.Contains(resp.Stdout, "pin-local ") {
			t.Fatalf("unexpected pin-local lines for no-matching-owner: %q", resp.Stdout)
		}
		t.Fatalf("expected already-up-to-date or applied 0 summary; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
}
```
