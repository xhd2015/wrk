## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <short hash>  initial status root
Status:       clean
Remote:       (no upstream)
```

## Expected

- Exit code 0.
- Stdout contains one status block for `Dir:          .`.
- The branch, short commit, and subject match git metadata.
- The status line is `clean`.
- Stderr is empty.

## Side Effects

- No repository files are changed.

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
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
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
