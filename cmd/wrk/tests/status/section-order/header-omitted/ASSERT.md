## Expected

- Exit code 0; stderr empty.
- Two primary blocks: main + out-of-tree linked (Master present).
- Stdout must **not** contain `---- external ----`.

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
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}
	assertNoExternalSectionHeader(t, resp.Stdout)
	assert.Output(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
		},
		nil,
	))
}
```
