## Expected

- Exit 0; stdout is the saved project absolute path + `\n`.
- Install hint; fake shell launched in saved path.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolvePath(t, req.MainRepo)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want+"\n"))
	assertInstallHint(t, resp.Stderr)
	assertFakeShellCwd(t, req, want)
}
```
