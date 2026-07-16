## Expected

- Flag layer accepts `--gen-commit-msg --commit --model=m --merge-back` together.
- Stderr must **not** contain `mutually exclusive`.
- Later-stage errors on main-repo cwd are acceptable for P2 flag-layer.

## Side Effects

- None required (flag-layer only).

## Exit Code

- Any, as long as failure is not flag mutual exclusion of this combo.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --gen-commit-msg --commit --merge-back as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
