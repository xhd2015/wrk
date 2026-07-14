## Expected

- Exit code 0.
- Stderr contains `worktree remove` and/or `merge` log line(s).
- Stderr matches verbose timestamp format.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "worktree remove") && !strings.Contains(resp.Stderr, "merge") {
		t.Fatalf("stderr should contain worktree remove and/or merge, got %q", resp.Stderr)
	}
	assertStderrVerboseFormat(t, resp.Stderr)
}
```