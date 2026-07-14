## Expected

- Exit code 0.
- Root block `Dir: .` includes `Remote:       identical`.
- Stderr is empty.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	block := statusRootBlockWithRemoteTemplate(t, req.MainRepo, "clean", "Remote:       identical")
	assert.Output(t, resp.Stdout, block)
}
```