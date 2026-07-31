## Expected Output

```
https://github.com/acme/app/pull/42
```

## Expected

- Exit code 0.
- Stdout is **URL only** + trailing newline (no create/attach tokens).
- Stderr empty.
- Fake `gh`: `pr list` called; **`pr create` and `pr comment` not called**.
- No `pushed` line (show never ensures/pushes).

## Side Effects

- No git push; no PR create; no issue comment.

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

	assert.Output(t, resp.Stdout, v2StdoutTemplate(prShowStdout(prDefaultURL)))
	for _, tok := range []string{"PR created", "comment added", "title set", "body set", "pushed "} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("show mode must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	_ = assertGhSubcmdCalled(t, invocs, "list")
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
