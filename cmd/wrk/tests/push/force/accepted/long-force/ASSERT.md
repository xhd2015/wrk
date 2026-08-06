## Expected Output

```
pushed main → origin/main
```

## Expected

- Exit code 0.
- Same confirm line as short `-f` (not `force-pushed`).
- Origin/main equals local HEAD.
- Stderr empty.

## Side Effects

- Branch tip published via force-with-lease.

## Exit Code

- 0

```go
import (
	"strings"

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
	if strings.Contains(resp.Stdout, "force-pushed") {
		t.Fatalf("confirm line must stay 'pushed …', not force-pushed; got %q", resp.Stdout)
	}

	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set by setupPushMainWithOrigin")
	}
	assertOriginBranchEqualsLocal(t, req.MainRepo, req.OriginBare, "main")
}
```
