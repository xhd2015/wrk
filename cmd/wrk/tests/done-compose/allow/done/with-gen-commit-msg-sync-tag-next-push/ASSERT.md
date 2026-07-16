## Expected

- Flag layer accepts `--gen-commit-msg --commit --model=m --done --sync --tag-next --push`.
- Stderr must **not** contain `mutually exclusive`.
- Stderr must **not** contain `--push is only valid with --tag-next`.
- Later-stage errors on main-repo cwd are acceptable for P2.

## Side Effects

- None required (flag-layer only).

## Exit Code

- Any, as long as failure is not flag-matrix rejection of this combo.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("--gen-commit-msg --commit --done --sync --tag-next --push still mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if strings.Contains(se, "--push is only valid with --tag-next") {
		t.Fatalf("push still tag-next-only with primary+gen-commit combo; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
