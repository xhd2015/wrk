## Expected Output

```
==== unwind (dry-run) ====
would: peel .
  would: generate commit message and commit staged changes
```

(No `would: leave N file(s) uncommitted …` line.)

## Expected

- Exit code 0.
- Peel display `.`.
- Stdout contains generate/commit plan language under peel.
- Stdout does **not** contain leave-uncommitted vocabulary.
- Stdout does **not** require `would: git add -A` (no `--add-all` flag).
- Zero mutations.

## Side Effects

- None (dry-run must not unstage or commit).

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
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")
	assertContainsInOrder(t, resp.Stdout,
		peelLine("."),
		"generate",
		"commit",
	)
	if strings.Contains(resp.Stdout, "uncommitted") ||
		(strings.Contains(resp.Stdout, "leave") && strings.Contains(resp.Stdout, "file")) {
		t.Fatalf("fully staged peel must omit leave-N line; stdout:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "would: git add -A") {
		t.Fatalf("without --add-all must not plan git add -A; stdout:\n%s", resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
