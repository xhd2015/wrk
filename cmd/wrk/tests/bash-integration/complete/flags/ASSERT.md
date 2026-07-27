
## Expected

- Exit code 0.
- Stdout contains multiple flag candidates; each line starts with `-`.
- Stdout includes key flags from usage: `--list`, `-l`, `--status`, `--dep`, `--where`, `--add`, `--rm`, `--bash-integration`.

## Side Effects

- Read-only completion.

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertCompleteExitOK(t, resp)
	assertAllLinesAreFlags(t, resp.Stdout)
	assertFlagsInclude(t, resp.Stdout,
		"--list",
		"-l",
		"--status",
		"--dep",
		"--where",
		"--add",
		"--rm",
		"--bash-integration",
	)
	assertNoEventsJSONL(t, resp)
}
```