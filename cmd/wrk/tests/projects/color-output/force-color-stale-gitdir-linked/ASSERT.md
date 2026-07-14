## Expected

- Exit code 0.
- `Worktrees:` shows `2 total, 0 dirty, ` uncolored and `1 error` in red ANSI.
- Per-worktree detail line has red on `error: <full git stderr>`.
- Stderr is empty.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertColorProjectsBlocksSeparated(t, resp.Stdout, 1)

	remote := colorCompareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	gitErr := colorWorktreeStatusError(t, req.WtDir)
	summary := colorFormatWorktreesSummary(2, 0, 1, 0)
	detail := colorWorktreeErrorDetailLine(t, req.WtDir, gitErr)
	block := colorProjectStatusBlockWithDetailsTemplate(t, req.MainRepo, "clean", remote, summary, []string{detail})
	assert.Output(t, resp.Stdout, block)
}
```