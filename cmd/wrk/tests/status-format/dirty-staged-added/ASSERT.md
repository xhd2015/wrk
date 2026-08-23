## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <short hash>  status format staged added base
Status:       dirty (1 staged, 0 changed, 0 renamed, 0 deleted, 0 untracked)
Remote:       (no upstream)
```

## Expected

- Exit code 0.
- Status is dirty with one **staged** from the staged new file.
- Zero untracked; all five buckets appear.

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
		"dirty (1 staged, 0 changed, 0 renamed, 0 deleted, 0 untracked)"))
}
```
