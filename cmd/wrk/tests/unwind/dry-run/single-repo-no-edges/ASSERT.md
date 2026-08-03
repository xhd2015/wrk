## Expected Output

```
==== unwind (dry-run) ====
would: peel .
```

## Expected

- Exit code 0.
- Exactly one peel step: `would: peel .` (primary checkout is cwd).
- Must **not** print bare MainRepo basename alone (`would: peel root`).
- No requirement for `--tag-next` / `--push` (command did not pass them; still succeeds).
- Zero mutations: HEAD unchanged; `DIRTY` file still present.

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
	// Single step: only one would: peel line.
	if n := strings.Count(resp.Stdout, "would: peel "); n != 1 {
		t.Fatalf("want exactly 1 peel line, got %d\nstdout:\n%s", n, resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
