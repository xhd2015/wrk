## Expected

- Non-zero exit (cannot resolve a remote for the current branch).
- Stderr explains missing upstream and/or origin (mentions `upstream`, `origin`, or `remote`).
- Stdout must **not** contain a success push line `pushed main → origin/main`.

## Errors

- Clear non-zero error when no remote can be resolved for the branch push.

## Exit Code

- Non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when --push has no remote; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "remote") && !strings.Contains(se, "origin") && !strings.Contains(se, "upstream") {
		t.Fatalf("stderr should explain missing remote/origin/upstream, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "pushed main → origin/main") {
		t.Fatalf("stdout must not claim successful push without a remote; got %q", resp.Stdout)
	}
}
```
