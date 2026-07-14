## Expected

- Exit code 0.
- `Remote:       needs pull(1 commit behind)`.
- Stderr is empty (no `-v`).

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
	remote := remoteFieldLine(t, req.MainRepo, "origin/main", "main")
	if remote != "Remote:       needs pull(1 commit behind)" {
		t.Fatalf("Remote: want needs pull(1 commit behind), got %q", remote)
	}
	block := projectsRemoteBlockTemplate(t, req.MainRepo, "clean", remote, "0 total, 0 dirty")
	assert.Output(t, resp.Stdout, block)
}
```