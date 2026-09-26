## Expected

- Exit code 0.
- Stdout contains mock title `feat: add feature`.
- HEAD exists (root commit) with subject `feat: add feature`.
- Stderr does not contain the unborn-HEAD `rev-parse` fatal.

## Side Effects

- First commit is created on the previously unborn branch.

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

	if strings.Contains(resp.Stderr, "ambiguous argument") ||
		strings.Contains(resp.Stderr, "unknown revision") {
		t.Fatalf("unborn HEAD must not fail exclusive-branch ReadBranch, stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "feat: add feature") {
		t.Fatalf("stdout missing title, got:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "git commit failed") {
		t.Fatalf("git commit should succeed, stderr:\n%s", resp.Stderr)
	}
	if !gitHEADExists(t, req.RepoDir) {
		t.Fatalf("expected root commit on previously unborn HEAD\nstdout=%q\nstderr=%q", resp.Stdout, resp.Stderr)
	}
	subject := gitHEADSubject(t, req.RepoDir)
	if subject != "feat: add feature" {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q", subject, "feat: add feature", resp.Stdout, resp.Stderr)
	}
}
```
