## Expected Output

```text
Dir:          ../..
Branch:       main
Commit:       <short hash>  add subdir file
Status:       clean
Remote:       (no upstream)
```

## Expected

- Exit code 0.
- Stdout reports main Dir as `../..` (Rel from `sub/dir`), **not** forced `.`.
- `Remote:` still present (main identity, not Dir==".").
- The branch, short commit, and subject match the root checkout.
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
	if got := statusOutputBlockCount(resp.Stdout); got != 1 {
		t.Fatalf("expected 1 status block, got %d:\n%s", got, resp.Stdout)
	}
	dir := statusDirLine(t, req.RepoDir, req.MainRepo)
	if dir != "../.." {
		t.Fatalf("fixture Dir want ../.., got %q", dir)
	}
	assert.Output(t, resp.Stdout, v2StdoutTemplate(statusMainBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean")))
}
```
