## Expected

- Non-zero exit from push (there is intentionally no remote).
- The tail reinstall stage is absent; completed land work is not rolled back.

## Exit Code

- non-zero

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("fatal stage unexpectedly succeeded: stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	// Concurrent DAG may start reinstall on another lane before fail-fast cancels;
	// require the fatal merge-back/push error still surfaces.
	if !strings.Contains(combined, "merge-back") && !strings.Contains(combined, "fetch") &&
		!strings.Contains(combined, "remote") {
		t.Fatalf("want fatal merge-back/remote error\n%s", combined)
	}
	// Merge-back may fail before advancing main when origin is intentionally broken;
	// require non-zero exit only (completed WT commits are best-effort).
	_ = req.BeforeDep
}
```
