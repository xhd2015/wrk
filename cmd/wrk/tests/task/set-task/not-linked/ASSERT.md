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
		t.Fatal("expected non-zero exit for --set-task from main repo")
	}
	// Should error about not being a linked worktree, NOT about terminal
	if !strings.Contains(resp.Stderr, "linked") && !strings.Contains(resp.Stderr, "worktree") {
		t.Fatalf("expected error mentioning linked worktree, got stderr=%q", resp.Stderr)
	}
	// Should NOT mention terminal (the linked check comes first)
	if strings.Contains(resp.Stderr, "terminal") || strings.Contains(resp.Stderr, "tty") {
		t.Fatalf("should not mention terminal before linked check fails, got stderr=%q", resp.Stderr)
	}
}
```
