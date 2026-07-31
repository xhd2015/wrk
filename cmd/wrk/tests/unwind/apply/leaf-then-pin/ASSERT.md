## Expected

- Exit code 0 (after implementer lands apply; RED while stubbed).
- Leaf main advanced with feature content; local tag `v0.0.2` at leaf main HEAD.
- Bare origin `main` matches leaf main HEAD; origin has `refs/tags/v0.0.2`.
- Root consumer `go.mod` requires `example.com/dot-pkgs v0.0.2` with **no** replace.
- Stdout/stderr may include peel/tag/push/pin progress (shape implementer-owned);
  must **not** be the apply-not-implemented stub once GREEN.

## Side Effects

- Leaf land merges branch into leaf main; tag-next + push publish tip + tag.
- Pin edits root module require (and tidy may touch go.sum).
- Leaf external path may be removed by `--done` (also covered by done-removes leaf).

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
	// Classic TDD: stub still fails apply → non-zero until implementer.
	// Assert side effects only on success so RED is the stub exit, not panics.
	if resp.ExitCode != 0 {
		combined := resp.Stdout + "\n" + resp.Stderr
		if strings.Contains(combined, "not implemented") {
			t.Fatalf("apply not implemented yet (expected RED until P4 lands): exit=%d stderr=%q stdout=%q",
				resp.ExitCode, resp.Stderr, resp.Stdout)
		}
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertLeafMainAdvancedAndTagged(t, req)
	assertConsumerPinned(t, req)
}
```
