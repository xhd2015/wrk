## Expected Output

```
==== unwind (dry-run) ====
would: peel .
  would: git add -A
  would: generate commit message and commit staged changes
```

(Indent of nested `would:` lines implementer-owned if consistent; asserts lock
substring order: peel → git add -A → generate/commit language. No leave-N line.)

## Expected

- Exit code 0.
- Peel display `.` (primary at cwd).
- Under the peel: `would: git add -A` appears **before** generate/commit plan language.
- Stdout does **not** contain leave-uncommitted vocabulary.
- Zero mutations (dry-run).

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
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")
	assertContainsInOrder(t, resp.Stdout,
		peelLine("."),
		"would: git add -A",
		"generate",
		"commit",
	)
	if strings.Contains(resp.Stdout, "leave") && strings.Contains(resp.Stdout, "uncommitted") {
		t.Fatalf("--add-all plan must not print leave-uncommitted; stdout:\n%s", resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
