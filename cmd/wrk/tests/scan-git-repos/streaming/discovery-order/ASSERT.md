## Expected

- Exit code 0.
- Stdout is exactly two lines (each absolute main path + trailing `\n` after last):
  1. `main-b` (first CLI root / discovery order)
  2. `main-a` (second CLI root)
- Order is **not** lexicographic (`main-a` then `main-b`).
- `projects.json` is empty (print-only).

## Side Effects

- No projects.json mutation.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	// Discovery order: first CLI root's main, then second (not lex path sort).
	wantFirst := resolveScanPath(t, req.MainRepo)
	wantSecond := resolveScanPath(t, req.SecondRepo)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(wantFirst+"\n"+wantSecond))

	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)
}
```
