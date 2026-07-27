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
		t.Fatal("expected non-zero exit for --set-task with empty-slug result")
	}
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "empty") && !strings.Contains(combined, "slug") {
		t.Fatalf("expected error about empty slug, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
