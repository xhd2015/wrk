## Expected

- Exit 0.
- Stderr contains `notice: worktree clean, skip commit`.
- HEAD subject unchanged.

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
	if !strings.Contains(resp.Stderr, "notice: worktree clean, skip commit") {
		t.Fatalf("expected clean-skip notice; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	subject := gitHEADSubject(t, req.RepoDir)
	if subject != req.HEADSubject {
		t.Fatalf("HEAD subject changed on soft-skip: before=%q after=%q", req.HEADSubject, subject)
	}
}
```
