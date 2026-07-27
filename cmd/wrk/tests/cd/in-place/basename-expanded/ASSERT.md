
## Expected

- Exit code 0.
- Stdout empty.
- Follow-up is `cd <expanded-saved-abs>\n` (not basename).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertFollowupCDLine(t, req, req.MainRepo)
	got := readFollowupFile(t, req.FollowupFile)
	if strings.Contains(got, "cd myrepo") {
		t.Fatalf("follow-up must use expanded abs, not basename; got %q", got)
	}
}
```
