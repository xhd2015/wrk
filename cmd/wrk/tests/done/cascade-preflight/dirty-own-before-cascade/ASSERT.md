## Expected

- Non-zero exit.
- Dirty/uncommitted language (own or cascade preflight framing with `Error:` preferred).
- **External still present** — proves preflight runs before cascade mutations (D2).
  Today cascade may remove external then fail on own dirty → **RED** until fixed.
- Consumer linked worktree still present.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit on dirty own preflight; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	combinedLower := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if !strings.Contains(combinedLower, "uncommitted") && !strings.Contains(combinedLower, "clean") &&
		!strings.Contains(combinedLower, "dirty") {
		t.Fatalf("expected dirty/uncommitted language; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}

	// Core D2 pin: cascade must not have run remove before own dirty fails.
	assertFileExists(t, req.ExternalWtDir)
	assertGitFileIsWorktreeLink(t, req.ExternalWtDir)
	assertWorktreeListContains(t, req.DepPath, req.ExternalWtDir)

	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```
