## Expected

- Exit code 0.
- Stdout is exactly main abs path + trailing `\n`.
- Stderr contains install hint `wrk --bash-integration --install`.
- Fake interactive shell launched with cwd = **main** (not linked wt).
- No bare-main already-at-root notice.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolvePath(t, req.MainRepo)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want+"\n"))
	assertInstallHint(t, resp.Stderr)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, want)
	assertNoBareMainNotice(t, resp.Stderr)
}
```
