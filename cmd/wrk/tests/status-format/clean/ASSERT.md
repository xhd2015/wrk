## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <short hash>  status format clean base
Status:       clean
Remote:       (no upstream)
```

## Expected

- Exit code 0.
- Stdout is one main-root status block.
- The status line is exactly `clean` (wrk-owned clean wording).
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
	_ = d
	assertStatusFormatOK(t, resp, err)
	assert.Output(t, resp.Stdout, statusFormatBlockTemplate(t, req.MainRepo, "clean"))
}
```
