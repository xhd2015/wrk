## Expected Output

Two absolute paths, lexicographically sorted, one per line, trailing newline after the last line.

## Expected

- Exit code 0.
- Stdout lists both saved absolute paths sorted lexicographically.
- Stderr is empty.

## Side Effects

- Read-only lookup; no worktrees created.

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

	sorted := sortedSavedPaths(t, req.MainRepo, req.SecondRepo)
	wantStdout := sorted[0] + "\n" + sorted[1] + "\n"
	assert.Output(t, resp.Stdout, v2StdoutTemplate(wantStdout))
}
```