## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <short hash>  <subject>
Status:       clean
Remote:       (no upstream)
```

## Expected

- Exit code 0.
- Stdout contains one status block for `Dir:          .` (saved repo root).
- Branch, commit, and status lines match the saved repo git metadata.
- Stderr is empty.

## Side Effects

- Status reported for the saved project path resolved via basename fallback, not cwd.

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
	if got := statusOutputBlockCount(resp.Stdout); got != 1 {
		t.Fatalf("expected 1 status block, got %d:\n%s", got, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, statusBlockTemplate(t, req.MainRepo, ".", "clean"))
}
```