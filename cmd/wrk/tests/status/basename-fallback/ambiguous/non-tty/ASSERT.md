## Expected

- Non-zero exit code.
- Stderr mentions multiple matches for `myrepo`.
- Stderr lists both candidate absolute paths (lexicographically sorted).
- Stdout is empty.

## Errors

- Ambiguous basename without TTY cannot proceed.

## Exit Code

- Non-zero

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}

	sorted := sortedSavedPaths(t, req.MainRepo, req.SecondRepo)
	tmpl := `<contains>
Multiple projects match "myrepo":
  1) ` + sorted[0] + `
  2) ` + sorted[1] + `
</contains>`
	assert.Output(t, resp.Stderr, tmpl)
}
```