## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <root short>  main ahead commit
Status:       clean

Dir:          wt-linked
Branch:       wt-side
Commit:       <wt short>  status main root
Status:       clean
Master:       needs fast forward(+1 commit)
```

## Expected

- Exit code 0.
- Stdout contains two status blocks: root `.` and linked worktree `wt-linked`.
- Root block has **no** `Master:` line.
- Linked worktree block includes one-line `Master: needs fast forward(+1 commit)`.
- Stderr is empty.

## Side Effects

- No repository files are changed.

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

	master := masterField(t, req.MainRepo, "main", req.WtBranch)
	assert.Output(t, resp.Stdout, statusStdoutV2(t,
		statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
		statusBlockWithMasterPlain(t, req.WtDir, "wt-linked", "clean", master),
	))
}
```