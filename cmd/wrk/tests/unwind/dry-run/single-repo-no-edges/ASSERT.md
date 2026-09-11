## Expected Output

```
would: peel .
```

(Phase banners on stderr; peel line on stdout when the action graph is empty.)

## Expected

- Exit code 0.
- Exactly one peel step: `would: peel .` (primary checkout is cwd).
- Must **not** print bare MainRepo basename alone (`would: peel root`).
- No requirement for `--tag-next` / `--push` (command did not pass them; still succeeds).
- Zero mutations: HEAD unchanged; `DIRTY` file still present.

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
	out := resp.Stdout
	assertPeelOrder(t, out, req.PeelOrder)
	assertPeelUsesRelDisplay(t, out, ".")
	if n := strings.Count(out, "would: peel "); n != 1 {
		t.Fatalf("want exactly 1 peel line, got %d\nstdout:\n%s", n, out)
	}
	assertUnwindZeroMutations(t, req)
}
```
