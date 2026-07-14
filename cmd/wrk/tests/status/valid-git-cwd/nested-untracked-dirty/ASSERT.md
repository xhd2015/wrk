## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <root short hash>  ignore nested tools
Status:       clean
Remote:       (no upstream)

Dir:          tools/child
Branch:       main
Commit:       <child short hash>  child status repo
Status:       dirty (1 added, 0 changed, 0 renamed, 0 deleted)
```

## Expected

- Exit code 0.
- Stdout contains one block for the root checkout as `.` with `Status: clean`.
- Stdout contains one block for the nested independent repository as `tools/child`.
- Nested block status is dirty with one **added** entry from the untracked file.
- Block order follows `scan_repo.Scan` path ordering.
- Stderr is empty.

## Side Effects

- The untracked file under `tools/child` remains untracked after status is printed.

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
		statusBlockPlain(t, req.DepPath, "tools/child", "dirty (1 added, 0 changed, 0 renamed, 0 deleted)"),
	))
}
```
