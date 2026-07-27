
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
		t.Fatal("expected non-zero exit for --task + --done")
	}
	if !strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("expected 'mutually exclusive' error, got stderr=%q", resp.Stderr)
	}
}
```
