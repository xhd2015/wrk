## Expected

- Non-zero exit (non-fast-forward rejected).
- Origin `refs/heads/main` unchanged from pre-run snapshot (remote-only tip).
- Stdout must **not** contain success line `pushed main → origin/main`.
- Local HEAD still the local-only tip (push did not rewrite local).

## Errors

- Non-FF push without force fails; origin preserved.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for non-FF --push without force; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "pushed main → origin/main") {
		t.Fatalf("must not claim successful push on non-FF without force; stdout=%q", resp.Stdout)
	}

	before := readOriginMainBefore(t, req)
	after := revParseRef(t, req.OriginBare, "refs/heads/main")
	if after != before {
		t.Fatalf("origin/main must stay at remote-only tip %s; got %s", before, after)
	}

	// Local tip must still differ from origin (diverged fixture preserved).
	local := revParseHEAD(t, req.MainRepo)
	if local == after {
		t.Fatal("expected local HEAD still != origin after failed non-FF push")
	}
}
```
