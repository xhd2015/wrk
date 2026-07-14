## Expected

- Exit code 0 (prunable worktrees counted in summary, not fatal).
- `Worktrees:    0 total, 0 dirty, 1 prune` — dead checkout excluded from total/dirty/error.
- No per-path prune detail lines.
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
	assertProjectsBlocksSeparated(t, resp.Stdout, 1)

	remote := compareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	summary := formatWorktreesSummary(0, 0, 0, 1)
	block := projectStatusBlockTemplate(t, req.MainRepo, "clean", remote, summary)
	assert.Output(t, resp.Stdout, block)
}
```