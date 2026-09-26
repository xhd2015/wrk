## Expected

- Exit code 0.
- Stdout is mock B for N=1.
- Stderr contains `would: git commit`.
- Stderr does not contain the unborn-HEAD `rev-parse` fatal.
- HEAD remains unborn (no commit created).

## Side Effects

- No new commit is created.
- Agent is not required (dry-run pure plan).

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
	assertMockMessageB(t, resp.Stdout, 1)

	if !strings.Contains(strings.ToLower(resp.Stderr), "would:") ||
		!strings.Contains(resp.Stderr, "git commit") {
		t.Fatalf("stderr should contain would: git commit plan, stderr:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run, stderr:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "ambiguous argument") ||
		strings.Contains(resp.Stderr, "unknown revision") {
		t.Fatalf("unborn HEAD must not fail exclusive-branch ReadBranch, stderr:\n%s", resp.Stderr)
	}
	if gitHEADExists(t, req.RepoDir) {
		t.Fatalf("HEAD must stay unborn under dry-run --commit")
	}
}
```
