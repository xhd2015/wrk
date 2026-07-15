## Expected

- Exit code 0.
- **No** `Remote:` anywhere in stdout.
- Linked-in-tree cwd path: first block is the current checkout as `Dir: .` (no Master);
  second block is the same linked path again with `Master:` — also `Dir: .` because
  `statusDirLine(invCwd=wt, path=wt)` is `.` (not main-relative `wt-linked`).
- Stderr is empty.

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

	master := masterIdenticalField(t, req.MainRepo, "main", "wt-from-linked")
	// Linked-in-tree shortcut: current cwd first (no Master), then all linked with Master.
	// Dir labels use invocation cwd → both blocks show "." for this single in-tree wt.
	rootBlock := fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       clean",
		statusBranchLine(t, req.WtDir), statusCommitLine(t, req.WtDir))
	wtBlock := fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       clean\n%s",
		statusBranchLine(t, req.WtDir), statusCommitLine(t, req.WtDir), master)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(joinStdoutBlocks(rootBlock, wtBlock)))
}
```
