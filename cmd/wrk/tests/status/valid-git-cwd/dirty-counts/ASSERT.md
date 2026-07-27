## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <short hash>  add dirty fixtures
Status:       dirty (1 added, 1 changed, 1 renamed, 1 deleted)
```

## Expected

- Exit code 0.
- Stdout reports the root checkout as `.`.
- Status is dirty with one added, one changed, one renamed, and one deleted entry.
- Stderr is empty.

## Side Effects

- Existing dirty worktree entries remain dirty after status is printed.

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
	assert.Output(t, resp.Stdout, statusBlockTemplate(t, req.MainRepo, ".", "dirty (1 added, 1 changed, 1 renamed, 1 deleted)"))
}
```
