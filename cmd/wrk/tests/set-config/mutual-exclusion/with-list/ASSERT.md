## Expected

- Non-zero exit.
- `--set-config` is recognized (not "unrecognized flag: --set-config").
- Stderr mentions mutual exclusion (or clearly rejects the flag combination).

## Errors

- Management mode cannot run alongside `--list`.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for mutual exclusion, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected mutual-exclusion error after --set-config is implemented; got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "mutually exclusive") || strings.Contains(resp.Stderr, "exclusive") {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}
```
