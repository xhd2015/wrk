
## Expected

- Exit code 0 despite failing pre-commit hook.
- Stderr does not report `git commit failed`.
- HEAD subject is exactly `feat: skip hooks`.

## Side Effects

- A new commit is created (hooks skipped via `--no-verify`).
- Contrast: the same repo with `--commit` only (no `--no-verify`) would fail at the hook
  (not a separate leaf; documented here).

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
		t.Fatalf("git commit should not fail with --no-verify, stderr:\n%s", resp.Stderr)
	}

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != "feat: skip hooks" {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q", subject, "feat: skip hooks", resp.Stdout, resp.Stderr)
	}
}
```
