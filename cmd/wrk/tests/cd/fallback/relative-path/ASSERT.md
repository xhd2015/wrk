
## Expected

- Exit 0; stdout absolute path of relative target + `\n`.
- Install hint on stderr; fake shell cwd = resolved abs.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolvePath(t, req.MainRepo)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want+"\n"))
	assertInstallHint(t, resp.Stderr)
	assertFakeShellCwd(t, req, want)
}
```
