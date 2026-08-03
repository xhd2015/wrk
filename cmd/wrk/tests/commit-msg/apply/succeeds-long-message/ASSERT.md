## Expected

- Exit code 0.
- HEAD subject is exactly `feat: long form`.

## Side Effects

- New commit with `--message` value as subject.

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
	if subject != "feat: long form" {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q", subject, "feat: long form", resp.Stdout, resp.Stderr)
	}
}
```
