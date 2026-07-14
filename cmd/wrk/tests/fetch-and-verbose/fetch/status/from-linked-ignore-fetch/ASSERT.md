## Expected

- Exit code 0.
- No `Remote:` anywhere in stdout.
- Linked block has `Master:` where applicable.
- Stderr is empty (no `-v`).

## Exit Code

- 0

```go
import (
	"fmt"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutNoRemoteField(t, resp.Stdout)

	master := masterIdenticalField(t, req.MainRepo, "main", "wt-branch")
	rootBlock := fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       clean",
		statusBranchLine(t, req.WtDir), statusCommitLine(t, req.WtDir))
	wtBlock := fmt.Sprintf("Dir:          wt-linked\n%s\n%s\nStatus:       clean\n%s",
		statusBranchLine(t, req.WtDir), statusCommitLine(t, req.WtDir), master)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(joinStdoutBlocks(rootBlock, wtBlock)))
}
```