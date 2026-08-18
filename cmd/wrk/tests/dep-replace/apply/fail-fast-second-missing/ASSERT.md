## Expected

- Non-zero exit (apply fail-fast).
- First dep replace is applied (sequential fail-fast leaves prior writes).
- Stderr indicates missing second path (`wrk:` + no such dir or equivalent).
- No success claim for the missing path.
- No complete success summary. If stdout leaked a first-success line, it must
  not be the old one-liner `dep-replace example.com/dep =>`.

## Side Effects

- Partial apply of first directory is allowed under D3 fail-fast.

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
	assertExitNonZero(t, resp)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertNotContains(t, resp.Stdout, "dep-replace "+modDep+" =>")
	assertNotContains(t, resp.Stdout, "dep-replace: replaced")
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "no such") &&
		!strings.Contains(se, "not found") &&
		!strings.Contains(se, "missing") &&
		!strings.Contains(se, "exist") &&
		!strings.Contains(se, "stat") {
		t.Fatalf("stderr should indicate second path missing, got %q", resp.Stderr)
	}
}
```
