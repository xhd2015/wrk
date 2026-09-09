## Expected Output

```text
Dir:          alpha
Branch:       main
Commit:       <alpha short hash>  alpha status repo
Status:       clean

Dir:          beta
Branch:       main
Commit:       <beta short hash>  beta status repo
Status:       clean
```

## Expected

- Exit code 0.
- Two status blocks in path-sorted order (`alpha` then `beta`).
- No `Remote:` lines and no `---- external ----` header.
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
	assertNoExternalSectionHeader(t, resp.Stdout)
	assert.Output(t, resp.Stdout, statusStdoutV2(t,
		statusBlockPlain(t, req.MainRepo, "alpha", "clean"),
		statusBlockPlain(t, req.DepPath, "beta", "clean"),
	))
}
```
