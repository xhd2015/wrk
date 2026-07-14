## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <root short hash>  root status repo
Status:       clean

Dir:          tools/child
Branch:       main
Commit:       <child short hash>  child status repo
Status:       clean
```

## Expected

- Exit code 0.
- Stdout contains one block for the root checkout as `.`.
- Stdout contains one block for the nested independent repository as `tools/child`.
- Block order follows `scan_repo.Scan` path ordering.
- Stderr is empty.

## Side Effects

- No repository files are changed.

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
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, statusStdoutV2(t,
		statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
		statusBlockPlain(t, req.DepPath, "tools/child", "clean"),
	))
}
```
