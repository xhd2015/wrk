## Expected

- Non-zero exit; empty stdout.
- Stderr lists both candidate absolute paths (lexicographically).

## Errors

- Ambiguous basename without TTY.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	sorted := sortedSavedPaths(t, req.MainRepo, req.SecondRepo)
	tmpl := `<contains>
Multiple projects match "myrepo":
  1) ` + sorted[0] + `
  2) ` + sorted[1] + `
</contains>`
	assert.Output(t, resp.Stderr, tmpl)
}
```
