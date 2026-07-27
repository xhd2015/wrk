## Expected

- Exit code 0.
- Help text (stdout and/or stderr) contains `--projects-dep-graph`.
- Prefer also documenting module-level dep graph / registered projects wording.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--projects-dep-graph") {
		t.Fatalf("help must mention --projects-dep-graph; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, help, `<contains>
--projects-dep-graph
</contains>`)
}
```
