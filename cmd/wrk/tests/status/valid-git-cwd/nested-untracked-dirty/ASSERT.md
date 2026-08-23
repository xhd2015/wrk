## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <root short hash>  ignore nested tools
Status:       clean
Remote:       (no upstream)

---- external ----

Dir:          tools/child
Branch:       main
Commit:       <child short hash>  child status repo
Status:       dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)
```

## Expected

- Exit code 0.
- Primary root block as `.` with `Status: clean` and Remote.
- Plain header `---- external ----` then nested `tools/child` block.
- Nested block status is dirty with one **untracked** entry from the untracked file.
- Stderr is empty.

## Side Effects

- The untracked file under `tools/child` remains untracked after status is printed.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
		},
		[]string{
			statusBlockPlain(t, req.DepPath, "tools/child", "dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)"),
		},
	))
}
```
