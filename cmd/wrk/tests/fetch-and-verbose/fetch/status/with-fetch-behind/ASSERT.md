## Expected

- Exit code 0.
- Root `Remote:       needs pull(1 commit behind)`.
- Stderr is empty.

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
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	remote := remoteFieldLine(t, req.MainRepo, "origin/main", "main")
	block := statusRootBlockWithRemoteTemplate(t, req.MainRepo, "clean", remote)
	assert.Output(t, resp.Stdout, block)
}
```