## Expected

- Exit code 0.
- Exactly one status block: `Dir: .` for the linked consumer (with `Master:`).
- No `---- external ----` section header.
- Stderr empty.

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
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if got := statusOutputBlockCount(resp.Stdout); got != 1 {
		t.Fatalf("expected 1 status block, got %d:\n%s", got, resp.Stdout)
	}
	assertNoExternalSectionHeader(t, resp.Stdout)

	assert.Output(t, resp.Stdout, statusStdoutV2(t,
		linkedScanBlock(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
	))
}
```
