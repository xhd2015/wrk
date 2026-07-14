---
label: slow
explanation: git init + linked worktree add; cold run ~15s
---

## Expected

- Exit code 0.
- Stdout contains the main repo path and the linked worktree path.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"path/filepath"

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

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	want := gitWorktreeListIsolated(t, mainRepo)
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch:\nwant:\n%q\ngot:\n%q", want, resp.Stdout)
	}
}
```