## Expected

- Non-zero exit code.
- Stderr contains `--github is only valid with --projects`.
- Stdout is empty.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
--github is only valid with --projects
</contains>`)
	if !strings.Contains(resp.Stderr, "github") {
		t.Fatalf("stderr should mention github, got %q", resp.Stderr)
	}
}
```
