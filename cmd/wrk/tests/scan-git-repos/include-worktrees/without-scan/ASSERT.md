## Expected

- Non-zero exit code.
- Stderr explains that `--include-worktrees` is only valid with `--scan-git-repos`.
- Stdout is empty.

## Errors

- `--include-worktrees` without `--scan-git-repos` is invalid (same family as bare `--no-cache`).

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for bare --include-worktrees, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
--include-worktrees is only valid with --scan-git-repos
</contains>`)
}
```
