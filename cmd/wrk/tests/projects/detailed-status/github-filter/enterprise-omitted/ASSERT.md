## Expected

- Exit code 0.
- Empty stdout (enterprise host is not github.com).
- Project remains in projects.json.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assertProjectsCount(t, req.WrkHome, 1)
}
```
