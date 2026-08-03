## Expected

- Exit code 0.
- Stderr contains a would-commit line (`would:` and `git commit`) and the message `feat: x`.
- HEAD subject is unchanged from before the run.
- Must not execute a real commit (no `Running git commit...` as executed path if that marker exists).

## Side Effects

- No new commit is created.

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

	se := resp.Stderr
	if !strings.Contains(strings.ToLower(se), "would:") ||
		!strings.Contains(se, "git commit") {
		t.Fatalf("stderr should contain would: git commit plan, stderr:\n%s", se)
	}
	if !strings.Contains(se, "feat: x") {
		t.Fatalf("stderr would-line should include message feat: x, stderr:\n%s", se)
	}
	if strings.Contains(se, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run, stderr:\n%s", se)
	}

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != req.HEADSubject {
		t.Fatalf("HEAD subject changed under dry-run: before=%q after=%q", req.HEADSubject, subject)
	}
}
```
