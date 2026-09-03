## Expected

- Non-zero exit.
- Stderr indicates no staged / nothing to commit.
- Stderr must **not** contain the soft-skip notice.
- HEAD subject unchanged.

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
	if strings.Contains(resp.Stderr, "notice: worktree clean, skip commit") {
		t.Fatalf("bare gen-commit must not soft-skip; stderr=%q", resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	ok := strings.Contains(errText, "nothing to commit") ||
		strings.Contains(errText, "no staged") ||
		strings.Contains(errText, "no changes")
	if !ok {
		t.Fatalf("expected nothing-to-commit error; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	subject := gitHEADSubject(t, req.RepoDir)
	if subject != req.HEADSubject {
		t.Fatalf("HEAD subject changed on fail: before=%q after=%q", req.HEADSubject, subject)
	}
}
```
