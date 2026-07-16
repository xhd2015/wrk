## Expected Output (stdout)

```
project good  (/abs/.../good)
  module example.com/good  dir=.

1 project  ·  1 module  ·  0 cross-edges
```

## Expected

- Exit code 0 (soft-skip, not hard failure).
- Stderr contains `warning:` and the missing absolute path.
- Stdout shows only the good project (no missing path as a project block).
- Footer counts only included projects/modules.

## Errors

- Soft only: missing path does not fail the command.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertMissingPathWarning(t, resp.Stderr, req.MissingPath)

	if strings.Contains(resp.Stdout, req.MissingPath) {
		t.Fatalf("stdout must not list missing project path %q, got %q", req.MissingPath, resp.Stdout)
	}

	var b strings.Builder
	b.WriteString(projectHeader(req.GoodPath))
	b.WriteByte('\n')
	b.WriteString(moduleLine("example.com/good", "."))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(graphFooter(1, 1, 0))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))
}
```
