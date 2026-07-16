## Expected

- Exit code 0.
- Stdout contains mock title `feat: add feature` (message printed before commit).
- HEAD subject is exactly `feat: add feature`.

## Side Effects

- A new commit is created with the generated message.
- Agent path is exercised via fake-opencode (no live LLM).

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)

	if !strings.Contains(resp.Stdout, "feat: add feature") {
		t.Fatalf("stdout missing title, got:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "git commit failed") {
		t.Fatalf("git commit should succeed, stderr:\n%s", resp.Stderr)
	}

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != "feat: add feature" {
		t.Fatalf("HEAD subject = %q, want %q\nstdout=%q\nstderr=%q", subject, "feat: add feature", resp.Stdout, resp.Stderr)
	}
}
```
