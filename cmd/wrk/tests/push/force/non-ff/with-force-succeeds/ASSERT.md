## Expected Output

```
pushed main → origin/main
```

## Expected

- Exit code 0.
- Stdout is the stable push confirmation (still `pushed …`, not `force-pushed`).
- Stderr empty.
- Origin `refs/heads/main` equals local HEAD (overwrote remote-only tip).
- Origin tip differs from pre-run remote-only snapshot.

## Side Effects

- Non-FF update of origin/main via `--force-with-lease`.

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
		t.Fatalf("confirm line must stay 'pushed …'; got %q", resp.Stdout)
	}

	before := readOriginMainBefore(t, req)
	assertOriginBranchEqualsLocal(t, req.MainRepo, req.OriginBare, "main")
	after := revParseRef(t, req.OriginBare, "refs/heads/main")
	if after == before {
		t.Fatalf("origin/main should have moved from remote-only tip %s to local HEAD", before)
	}
}
```
