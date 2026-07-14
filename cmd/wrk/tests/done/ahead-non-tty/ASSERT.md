## Expected

- Non-zero exit code.
- Stderr mentions `stdin is not a terminal` or `cannot prompt`.
- Worktree directory still exists.
- Main repo does NOT have the worktree commit.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	combined := resp.Stderr + resp.Stdout
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "stdin is not a terminal") && !strings.Contains(lower, "cannot prompt") {
		t.Fatalf("expected confirmation error in output, got stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}

	assertFileExists(t, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))
}
```