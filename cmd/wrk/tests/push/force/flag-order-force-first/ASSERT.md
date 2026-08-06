## Expected Output

```
pushed main → origin/main
```

## Expected

- Exit code 0.
- Same confirm line and origin update as `--push -f`.
- Stderr empty.

## Side Effects

- Branch tip published; force flag position does not change semantics.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	assert.Output(t, resp.Stdout, v2StdoutTemplate(pushConfirmLine("main")))
	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set")
	}
	assertOriginBranchEqualsLocal(t, req.MainRepo, req.OriginBare, "main")
}
```
