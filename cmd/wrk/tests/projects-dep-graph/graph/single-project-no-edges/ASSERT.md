## Expected Output

```
project solo  (/abs/path/to/solo)
  module example.com/solo  dir=.

1 project  ·  1 module  ·  0 cross-edges
```

## Expected

- Exit code 0.
- One project header with basename `solo` and absolute path.
- One module line for `example.com/solo` at `dir=.`.
- No cross-edge arrow lines (`→`).
- Footer `1 project  ·  1 module  ·  0 cross-edges`.
- Stderr empty.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "→") {
		t.Fatalf("single project must not print cross-edge arrows, got %q", resp.Stdout)
	}
	var b strings.Builder
	b.WriteString(projectHeader(req.SoloPath))
	b.WriteByte('\n')
	b.WriteString(moduleLine("example.com/solo", "."))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(graphFooter(1, 1, 0))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))
}
```
