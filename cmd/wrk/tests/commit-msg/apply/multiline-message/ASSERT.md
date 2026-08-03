## Expected

- Exit code 0.
- HEAD subject (first line) is exactly `feat: subj`.

## Side Effects

- New commit; body may be present in full message (subject assert is load-bearing).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != "feat: subj" {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q", subject, "feat: subj", resp.Stdout, resp.Stderr)
	}
}
```
