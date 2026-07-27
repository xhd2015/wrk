## Expected

- Non-zero exit code (hard error; not a partial skip).
- Stderr mentions detached HEAD and/or that main is not on a named branch.
- Stdout does **not** print a successful `synced:` summary line.
- HEAD still detached at the same commit.

## Exit Code

- Non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for detached main, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	combined := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(combined, "detached") &&
		!strings.Contains(combined, "named branch") &&
		!strings.Contains(combined, "not on a branch") {
		t.Fatalf("expected detached/named-branch error, got stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "synced:") {
		t.Fatalf("stdout should not include successful synced summary, got %q", resp.Stdout)
	}

	assertHEADUnchanged(t, req.MainRepo, req.MainSHA)
}
```
