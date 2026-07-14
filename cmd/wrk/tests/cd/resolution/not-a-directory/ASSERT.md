## Expected

- Non-zero exit; empty stdout.
- Stderr indicates the path is not a directory (or cannot be used as cd target).

## Errors

- Target exists but is not a directory.

## Exit Code

- Non-zero

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	// Accept either explicit "not a directory" or a generic path error naming the file.
	if !strings.Contains(resp.Stderr, "not a directory") && !strings.Contains(resp.Stderr, "not a dir") {
		assert.Output(t, resp.Stderr, `<contains>
not a directory
</contains>`)
	}
}
```
