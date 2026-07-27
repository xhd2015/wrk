## Expected

- Exit code 0.
- `.gitignore` contains exactly one `/external` line (no duplicate appended).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	n, err := countGitignoreExternalLines(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if n != 1 {
		t.Fatalf(".gitignore should have exactly 1 /external line, got %d", n)
	}
}
```