## Expected

- Exit code 0.
- HEAD subject is exactly `feat: x`.
- Stderr does not report `git commit failed`.

## Side Effects

- A new commit is created with the user-supplied message (no AI).

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

	if strings.Contains(resp.Stderr, "git commit failed") {
		t.Fatalf("git commit should succeed, stderr:\n%s", resp.Stderr)
	}

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != "feat: x" {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q", subject, "feat: x", resp.Stdout, resp.Stderr)
	}
}
```
