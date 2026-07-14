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
		t.Fatal(`expected non-zero exit for wrk <dir> --set-task ""`)
	}
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "empty") && !strings.Contains(combined, "task") {
		t.Fatalf("expected error about empty task/description, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Worktree should NOT have been moved
	assertFileExists(t, req.WtDir)
}
```