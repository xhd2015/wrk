## Expected

- Exit code 0.
- Stdout has no ANSI escape sequences.
- Linked worktree shows plain one-line `Master:       needs fast forward(+1 commit)`.
- Stderr is empty.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertNoANSI(t, resp.Stdout)
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}

	master := colorStatusMasterFieldPlain(t, req.MainRepo, "main", req.WtBranch)
	assert.Output(t, resp.Stdout, statusStdoutV2(t,
		statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
		colorStatusBlockPlain(t, req.WtDir, "wt-linked", "clean", master),
	))
}
```