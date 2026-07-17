```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatal("expected non-zero exit for --set-task --confirm without TTY")
	}
	if !strings.Contains(resp.Stderr, "terminal") && !strings.Contains(resp.Stderr, "tty") {
		t.Fatalf("expected error about terminal/tty, got stderr=%q", resp.Stderr)
	}
	// Worktree should NOT have been moved (original path still exists)
	assertFileExists(t, req.WtDir)
}
```
