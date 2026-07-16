## Expected

- Non-zero exit code.
- Stderr states that `--no-verify` requires `--commit` (mentions both flags).

## Side Effects

- No agent invocation; pure flag validation.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	errText := resp.Stderr
	if !strings.Contains(errText, "--no-verify") || !strings.Contains(errText, "--commit") {
		t.Fatalf("stderr should mention --no-verify requires --commit, got %q", errText)
	}
	assert.Output(t, resp.Stderr, `<contains>
--no-verify
</contains>`)
}
```
