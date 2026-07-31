## Expected Output

```
wrk: ... cycle ...
```

(Exact wording implementer-owned; must include the substring `cycle`.)

## Expected

- Exit code non-zero.
- Stderr and/or stdout mentions `cycle`.
- Must **not** be solely the apply-not-implemented stub without cycle mention.
- No multi-step successful peel plan.
- Zero mutations: host wt + both externals still present; HEADs match baseline.

## Side Effects

- None (preflight abort).

## Exit Code

- non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertCycleError(t, resp)
	// Prefer cycle preflight over apply stub: combined must mention cycle
	// (assertCycleError already checks). Guard against silent apply path.
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if strings.Contains(combined, "not implemented") && !strings.Contains(combined, "cycle") {
		t.Fatalf("apply-mode cycle must fail on cycle preflight, not only apply stub; stderr=%q stdout=%q",
			resp.Stderr, resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
