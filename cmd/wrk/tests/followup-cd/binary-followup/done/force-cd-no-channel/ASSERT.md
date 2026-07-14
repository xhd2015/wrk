## Expected

- Exit code 0.
- Stdout contains `worktree removed:` (done contract unchanged).
- Operated worktree B is gone.
- Stderr contains install hint `wrk --bash-integration --install`.
- Fake interactive shell launched with cwd = main repo.
- Follow-up file empty (channel closed).

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertContains(t, resp.Stdout, "worktree removed:")
	assertFileNotExists(t, req.WtDir)
	assertInstallHint(t, resp.Stderr)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, req.MainRepo)
	assertFollowupEmpty(t, resp)
}
```
