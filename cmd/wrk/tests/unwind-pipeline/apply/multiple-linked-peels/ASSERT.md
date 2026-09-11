## Expected

- Exit 0 and gen-commit once for each dirty linked worktree, not once globally.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertExit0(t, resp)
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	// Phase progress labels gen-commit-msg (legacy said "generate").
	n := strings.Count(combined, "gen-commit-msg")
	if n < 2 {
		n = strings.Count(combined, "generate")
	}
	if n < 2 {
		t.Fatalf("want per-peel gen-commit >=2, got %d: stdout=%q stderr=%q", n, resp.Stdout, resp.Stderr)
	}
}
```
