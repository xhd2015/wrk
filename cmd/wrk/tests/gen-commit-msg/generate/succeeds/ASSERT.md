
## Expected

- Exit code 0.
- Stdout contains mock title `feat: add feature`.
- Stdout contains mock description `Implement feature X`.
- HEAD subject is unchanged (no `--commit`).

## Side Effects

- Agent path is exercised via fake-opencode (no live LLM).
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

	if !strings.Contains(resp.Stdout, "feat: add feature") {
		t.Fatalf("stdout missing title, got:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Implement feature X") {
		t.Fatalf("stdout missing description, got:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != req.HEADSubject {
		t.Fatalf("HEAD subject changed without --commit: before=%q after=%q", req.HEADSubject, subject)
	}
}
```
