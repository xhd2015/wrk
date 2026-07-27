
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
		t.Fatal("expected non-zero exit for --set-task on non-wrk worktree")
	}
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "branch") && !strings.Contains(combined, "pattern") && !strings.Contains(combined, "parse") {
		t.Fatalf("expected error about branch name/pattern, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
