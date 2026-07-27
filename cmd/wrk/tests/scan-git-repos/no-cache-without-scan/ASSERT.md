## Expected

- Non-zero exit code.
- Stderr explains that `--no-cache` is only valid with `--scan-git-repos`.
- Stdout is empty.

## Errors

- `--no-cache` without `--scan-git-repos` is invalid (same family as `--port` without `--web`, `--fetch` without `--projects`/`--status`).

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
		t.Fatalf("expected non-zero exit for bare --no-cache, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
--no-cache is only valid with --scan-git-repos
</contains>`)
}
```
