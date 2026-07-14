## Expected

- Non-zero exit.
- Flags are accepted (not "unrecognized flag").
- Stderr explains window requires terminal / conflict with `--no-new-terminal`.
- Prefer no successful space/iterm UX.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --new-window --no-new-terminal, stdout=%q", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected validation error after flags are implemented; got %q", resp.Stderr)
	}
	// Must mention both axes or the no-new-terminal constraint.
	if strings.Contains(resp.Stderr, "no-new-terminal") ||
		(strings.Contains(resp.Stderr, "window") && strings.Contains(resp.Stderr, "terminal")) {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
no-new-terminal
</contains>`)
}
```
