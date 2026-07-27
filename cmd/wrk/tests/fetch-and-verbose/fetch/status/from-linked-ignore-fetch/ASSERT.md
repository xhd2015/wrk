## Expected

- Exit code 0.
- No `Remote:` anywhere in stdout.
- Linked block has `Master:` where applicable; Dir for the current linked cwd is `.`
  (statusDirLine), not main-relative `wt-linked`.
- Stderr is empty (no `-v`).

## Exit Code

- 0

```go
import (
	"fmt"

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
	assertStdoutNoRemoteField(t, resp.Stdout)

	master := masterIdenticalField(t, req.MainRepo, "main", "wt-branch")
	// Same linked-in-tree layout as from-linked-no-remote: Dir relative to inv cwd.
	rootBlock := fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       clean",
		statusBranchLine(t, req.WtDir), statusCommitLine(t, req.WtDir))
	wtBlock := fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       clean\n%s",
		statusBranchLine(t, req.WtDir), statusCommitLine(t, req.WtDir), master)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(joinStdoutBlocks(rootBlock, wtBlock)))
}
```
