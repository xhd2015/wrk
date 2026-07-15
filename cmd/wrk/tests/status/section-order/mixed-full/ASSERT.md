## Expected

- Exit code 0; stderr empty.
- Four `Dir:` blocks: main + two ListLinked primary linked + nested external.
- Plain `---- external ----` between last primary and nested.
- Primary linked order matches `worktree.ListLinked` porcelain order.

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
	if got := statusOutputBlockCount(resp.Stdout); got != 4 {
		t.Fatalf("expected 4 status blocks, got %d:\n%s", got, resp.Stdout)
	}

	linked := listLinkedPaths(t, req.MainRepo)
	if len(linked) != 2 {
		t.Fatalf("expected 2 ListLinked entries, got %d: %v", len(linked), linked)
	}

	primary := []string{
		statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
	}
	for _, p := range linked {
		branch := req.WtBranch
		if resolvePath(t, p) == resolvePath(t, req.DepsLinkedWtDir) {
			branch = req.Wt2Branch // in-tree branch "wt-side"
		}
		primary = append(primary, primaryLinkedBlockPlain(t, req.RepoDir, req.MainRepo, p, branch, "clean"))
	}

	assert.Output(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		primary,
		[]string{
			statusBlockPlain(t, req.DepPath, "tools/child", "clean"),
		},
	))
}
```
