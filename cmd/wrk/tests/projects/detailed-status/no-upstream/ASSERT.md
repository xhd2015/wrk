## Expected

- Exit code 0.
- `Remote:       (no upstream)`.
- `Worktrees:    0 total, 0 dirty`.
- Stderr is empty.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertProjectsBlocksSeparated(t, resp.Stdout, 1)

	block := projectStatusBlockTemplate(t, req.MainRepo, "clean", "Remote:       (no upstream)", "0 total, 0 dirty")
	assert.Output(t, resp.Stdout, block)
}
```