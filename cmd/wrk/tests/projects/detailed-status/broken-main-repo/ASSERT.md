## Expected

- Exit code 0 (broken main repo surfaces inline error, not fatal).
- Block contains only `Dir:` and `Status:       error: <full git stderr>`.
- No `Branch:`, `Commit:`, `Remote:`, or `Worktrees:` lines.
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

	gitErr := mainRepoStatusError(t, req.MainRepo)
	block := brokenMainRepoBlockTemplate(t, req.MainRepo, gitErr)
	assert.Output(t, resp.Stdout, block)
}
```