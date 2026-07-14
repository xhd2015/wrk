## Expected Output

- Stderr contains `--repos is mutually exclusive`.

## Expected

- Exit code is non-zero.
- Stdout is empty.

## Side Effects

- No repository files are changed.

## Exit Code

- non-zero

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "--repos is mutually exclusive") {
		t.Fatalf("stderr should contain %q, got %q", "--repos is mutually exclusive", resp.Stderr)
	}
}
```
