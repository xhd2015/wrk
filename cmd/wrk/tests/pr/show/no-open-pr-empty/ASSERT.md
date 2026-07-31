## Expected

- Exit code 0.
- Stdout is **empty** (no bytes).
- Stderr empty (no required warning when no open PR).
- Fake `gh`: **`pr create` and `pr comment` not called**; `pr list` may be called.
- No create/attach tokens and no `pushed` line.

## Side Effects

- No git push; no PR create; no issue comment.

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout != "" {
		t.Fatalf("show with no open PR must have empty stdout; got %q", resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	for _, tok := range []string{"PR created", "comment added", "title set", "body set", "pushed "} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
	// list is allowed (and expected after implementer wires show); not required pre-impl.
	_ = invocs
}
```
