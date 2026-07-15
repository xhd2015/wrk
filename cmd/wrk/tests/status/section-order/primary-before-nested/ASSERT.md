## Expected

- Exit code 0; stderr empty.
- Three `Dir:` blocks total.
- Order: main primary, WRK linked primary, then plain header, then nested `aaa`.
- Proves primary-before-external vs legacy scan-then-append (`main`, `aaa`, then append wt).

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

	assert.Output(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
		},
		[]string{
			statusBlockPlain(t, req.DepPath, "aaa", "clean"),
		},
	))
}
```
