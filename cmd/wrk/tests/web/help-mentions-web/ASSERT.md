## Expected

- Exit code 0.
- Help text (stdout and/or stderr) contains `--web`.
- Help text also contains `--port`.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--web") {
		t.Fatalf("help must mention --web; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(help, "--port") {
		t.Fatalf("help must mention --port; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Prefer stdout for usage (current wrk -h prints to stdout).
	assert.Output(t, help, `<contains>
--web
</contains>`)
}
```
