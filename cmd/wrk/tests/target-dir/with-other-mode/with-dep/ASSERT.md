## Expected

- Non-zero exit code.
- Stderr contains `wrk: unexpected arguments` (target-dir is create-only; `--dep` rejects it).
- Stdout is empty.
- No external dep worktree is spawned under `{consumerTop}/external/`.
- Consumer `go.mod` is NOT modified (no `replace` added for the dep).

## Errors

- `<target-dir>` combined with `--dep`.

## Exit Code

- Non-zero

> Note: this leaf is a regression guard for the create-only contract. The current
> implementation already rejects any extra positional alongside `--dep` with
> `wrk: unexpected arguments`, so it may pass before the target-dir feature lands; it
> locks in the behavior so a future change that consumes `<target-dir>` in the dep path
> would be caught.

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
wrk: unexpected arguments
</contains>`)

	// no external dep worktree spawned
	assertFileNotExists(t, filepath.Join(req.TargetDir, "external"))

	// consumer go.mod should not have gained a replace directive
	data, readErr := os.ReadFile(filepath.Join(req.TargetDir, "go.mod"))
	if readErr != nil {
		t.Fatalf("read go.mod: %v", readErr)
	}
	if strings.Contains(string(data), "replace ") {
		t.Fatalf("go.mod should not contain a replace directive, got:\n%s", data)
	}
}
```
