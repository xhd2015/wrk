---
label: slow
---
## Expected

- Exit code 0 (stale-gitdir linked worktrees surface as inline errors, not fatal).
- Main project block is healthy with `Worktrees:    2 total, 0 dirty, 1 error`.
- Detail line for `stale-wt` shows full git stderr after `error:`.
- Pipe mode without `--color` — plain text, no ANSI.
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
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertProjectsBlocksSeparated(t, resp.Stdout, 1)

	remote := compareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	gitErr := worktreeStatusError(t, req.WtDir)
	summary := formatWorktreesSummary(2, 0, 1, 0)
	detail := worktreeErrorDetailLine(t, req.WtDir, gitErr)
	block := projectStatusBlockWithDetailsTemplate(t, req.MainRepo, "clean", remote, summary, []string{detail})
	assert.Output(t, resp.Stdout, block)
}
```