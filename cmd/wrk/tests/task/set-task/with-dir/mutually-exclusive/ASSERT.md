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
		t.Fatal("expected non-zero exit for wrk <dir> --set-task --list")
	}
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "mutually exclusive") {
		t.Fatalf("expected mutually exclusive error, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Worktree should NOT have been moved
	assertFileExists(t, req.WtDir)
}
```