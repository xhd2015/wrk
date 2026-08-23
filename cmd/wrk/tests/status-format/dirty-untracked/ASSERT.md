## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <short hash>  status format dirty base
Status:       dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)
Remote:       (no upstream)
```

## Expected

- Exit code 0.
- Stdout is one main-root status block.
- Status is dirty with one **untracked** from the untracked file (`??` → wrk `untracked`).
- Zero added, changed, renamed, and deleted; all five buckets appear in the dirty string.
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
	_ = d
	assertStatusFormatOK(t, resp, err)
	assert.Output(t, resp.Stdout, statusFormatBlockTemplate(t, req.MainRepo,
		"dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)"))
}
```
