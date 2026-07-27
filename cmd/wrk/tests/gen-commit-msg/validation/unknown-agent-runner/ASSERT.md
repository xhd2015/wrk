## Expected

- Non-zero exit code.
- Stderr mentions unsupported agent runner and `codex` (library pattern).

## Side Effects

- Mock success path is not taken (stdout is not mock B alone as success).
- Agent is not invoked.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	combined := resp.Stderr + "\n" + resp.Stdout
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "unsupported agent runner") {
		// Allow split wording across "unsupported" + "agent runner".
		if !(strings.Contains(lower, "unsupported") && strings.Contains(lower, "agent runner")) {
			t.Fatalf("stderr should mention unsupported agent runner, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
		}
	}
	if !strings.Contains(combined, "codex") {
		t.Fatalf("error should mention codex, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Successful dry-run mock must not be the only outcome.
	if resp.Stdout == mockMessageB(1) && strings.TrimSpace(resp.Stderr) == "" {
		t.Fatalf("should not succeed with mock B for unsupported runner")
	}
	assert.Output(t, resp.Stderr, `<contains>
codex
</contains>`)
}
```