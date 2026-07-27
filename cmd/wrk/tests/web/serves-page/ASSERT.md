---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Root probe performed HTTP GET `/`.
- `HTTPStatus` is 200.
- `HTTPBody` is HTML and includes workflow markers: `task`, `Main`, `Remote`,
  `wrk`, and `worktree` (or `worktrees` / `changes`).
- Stdout is exactly one listen URL line with trailing `\n`:
  `http://127.0.0.1:<port>/\n`.

## Side Effects

- Server process is terminated by the harness after the probe (no suite hang).

## Exit Code

- Ignored after kill (non-zero from SIGTERM is OK).

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.HTTPStatus != 200 {
		t.Fatalf("GET / expected 200, got %d body=%q stdout=%q stderr=%q",
			resp.HTTPStatus, resp.HTTPBody, resp.Stdout, resp.Stderr)
	}
	body := resp.HTTPBody
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "<html") && !strings.Contains(lower, "<!doctype") {
		t.Fatalf("expected HTML body, got %q", body)
	}
	for _, marker := range []string{"task", "Main", "Remote", "wrk"} {
		if !strings.Contains(body, marker) {
			t.Fatalf("HTML body missing marker %q; body=%q", marker, body)
		}
	}
	if !strings.Contains(body, "worktree") && !strings.Contains(body, "worktrees") && !strings.Contains(body, "changes") {
		t.Fatalf("HTML body missing worktree/changes marker; body=%q", body)
	}
	// Exact single-line listen URL with trailing newline (port is free/dynamic).
	assert.Output(t, resp.Stdout, `---
version: 3
__PORT__: type=number, example=18080, TCP listen port
---
http://127\.0\.0\.1:__PORT__/
`)
}
```
