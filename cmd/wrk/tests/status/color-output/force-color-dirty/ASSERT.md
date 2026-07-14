## Expected

- Exit code 0.
- `Status:` value uses granular coloring: red `dirty` and non-zero count segments, grey zero-count segments.
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
	if got := statusOutputBlockCount(resp.Stdout); got != 1 {
		t.Fatalf("expected 1 status block, got %d:\n%s", got, resp.Stdout)
	}

	status := colorStatusFormatDirtyCounts(1, 1, 0, 0)
	block := colorStatusBlockTemplate(t, req.MainRepo, ".", status, "")
	assert.Output(t, resp.Stdout, block)
}
```