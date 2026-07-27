```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
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