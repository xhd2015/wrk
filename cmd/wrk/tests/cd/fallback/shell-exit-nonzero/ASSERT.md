
## Expected

- Exit code 42 (propagated from fake shell).
- Stdout still printed abs path before shell wait.
- Install hint on stderr.
- Fake shell launched in target dir.

## Exit Code

- 42

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 42 {
		t.Fatalf("expected exit 42 from shell, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolvePath(t, req.MainRepo)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want+"\n"))
	assertInstallHint(t, resp.Stderr)
	assertFakeShellCwd(t, req, want)
}
```
