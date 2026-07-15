## Expected

- Exit code 0; stderr empty.
- Three primary blocks: main then two ListLinked paths in porcelain order.
- No `---- external ----` header.

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
	if got := statusOutputBlockCount(resp.Stdout); got != 3 {
		t.Fatalf("expected 3 status blocks, got %d:\n%s", got, resp.Stdout)
	}
	assertNoExternalSectionHeader(t, resp.Stdout)

	order := listLinkedPaths(t, req.MainRepo)
	if len(order) != 2 {
		t.Fatalf("expected 2 ListLinked entries, got %d: %v", len(order), order)
	}
	if resolvePath(t, order[0]) != resolvePath(t, req.WtDir) ||
		resolvePath(t, order[1]) != resolvePath(t, req.Wt2Dir) {
		t.Fatalf("ListLinked order: want [%q, %q], got %v", req.WtDir, req.Wt2Dir, order)
	}

	assert.Output(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.Wt2Dir, req.Wt2Branch, "clean"),
		},
		nil,
	))
}
```
