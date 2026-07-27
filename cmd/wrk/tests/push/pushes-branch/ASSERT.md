
## Expected Output

```
pushed main → origin/main
```

## Expected

- Exit code 0.
- Stdout exactly the stable push confirmation for main.
- Stderr empty.
- Bare origin `refs/heads/main` equals local main HEAD.

## Side Effects

- Branch tip published to origin via `runPushMain` semantics (branch-only, no tags required).

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
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	assert.Output(t, resp.Stdout, v2StdoutTemplate(pushConfirmLine("main")))

	if req.OriginBare == "" {
		t.Fatal("OriginBare must be set by setupPushMainWithOrigin")
	}
	assertOriginBranchEqualsLocal(t, req.MainRepo, req.OriginBare, "main")
}
```
