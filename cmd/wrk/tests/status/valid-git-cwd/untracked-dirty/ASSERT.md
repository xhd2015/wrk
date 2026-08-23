## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <short hash>  untracked dirty base
Status:       dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)
Remote:       (no upstream)
```

## Expected

- Exit code 0.
- Stdout reports the root checkout as `.`.
- Status is dirty with one **untracked** entry from the untracked file (`??` maps to wrk `untracked`).
- Zero changed, renamed, and deleted.
- Stderr is empty.

## Side Effects

- The untracked file remains untracked after status is printed.

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
	assert.Output(t, resp.Stdout, statusBlockTemplate(t, req.MainRepo, ".", "dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)"))
}
```
