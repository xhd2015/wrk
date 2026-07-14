## Expected

- Exit code 0.
- Stderr contains `fetch` verbose log line with timestamp format.
- Stdout root block includes `Remote:       identical`.
- Stdout unchanged aside from new `Remote:` field.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertStderrContainsGitSubcommand(t, resp.Stderr, "fetch")
	assertStderrVerboseFormat(t, resp.Stderr)
	block := statusRootBlockWithRemoteTemplate(t, req.MainRepo, "clean", "Remote:       identical")
	assert.Output(t, resp.Stdout, block)
}
```