## Expected

- Exit code 0.
- Stdout is mock B for N=1.
- Stderr contains a would-commit line (`would: git commit`).
- HEAD subject is unchanged from before the run.

## Side Effects

- No new commit is created.
- Agent is not required (dry-run pure plan).

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertMockMessageB(t, resp.Stdout, 1)

	if !strings.Contains(strings.ToLower(resp.Stderr), "would:") ||
		!strings.Contains(resp.Stderr, "git commit") {
		t.Fatalf("stderr should contain would: git commit plan, stderr:\n%s", resp.Stderr)
	}
	// Real commit path markers must not appear as an executed commit.
	if strings.Contains(resp.Stderr, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run, stderr:\n%s", resp.Stderr)
	}

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != req.HEADSubject {
		t.Fatalf("HEAD subject changed under dry-run --commit: before=%q after=%q", req.HEADSubject, subject)
	}
}
```
