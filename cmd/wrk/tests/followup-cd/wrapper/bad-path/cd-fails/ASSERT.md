---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Non-zero exit (wrapper fails even though stub binary exits 0).
- Stderr contains `cd <missing-path>` (printed before failed builtin cd).
- FinalPWD remains start dir (cd failed).

## Exit Code

- Non-zero

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero wrapper exit when cd fails; stderr=%q stdout=%q pwd=%q",
			resp.Stderr, resp.Stdout, resp.FinalPWD)
	}
	wantCD := "cd " + req.FakeFollowupCD
	assertContains(t, resp.Stderr, wantCD)
	assertPathsEqual(t, resp.FinalPWD, req.StartDir)
}
```
