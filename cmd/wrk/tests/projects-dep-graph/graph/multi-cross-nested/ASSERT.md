## Expected Output

```
project app  (/abs/.../app)
  module example.com/app  dir=.
    → example.com/lib@v1.2.3  [lib]

project lib  (/abs/.../lib)
  module example.com/lib  dir=.
  module example.com/lib/sub  dir=sub

2 projects  ·  3 modules  ·  1 cross-edge
```

## Expected

- Exit code 0.
- Projects printed in sorted path order (`app` before `lib` under `repos/`).
- `lib` lists root + nested `sub` modules.
- Under `app` module: one cross-edge `→ example.com/lib@v1.2.3  [lib]`.
- No arrow for `example.com/external` (unknown owner).
- Footer `2 projects  ·  3 modules  ·  1 cross-edge`.
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
	if strings.Contains(resp.Stdout, "example.com/external") {
		t.Fatalf("external (unregistered) require must not appear in graph output, got %q", resp.Stdout)
	}
	// Path sort: .../repos/app before .../repos/lib
	var b strings.Builder
	b.WriteString(projectHeader(req.AppPath))
	b.WriteByte('\n')
	b.WriteString(moduleLine("example.com/app", "."))
	b.WriteByte('\n')
	b.WriteString(crossEdgeLine("example.com/lib", "v1.2.3", req.LibPath))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(projectHeader(req.LibPath))
	b.WriteByte('\n')
	b.WriteString(moduleLine("example.com/lib", "."))
	b.WriteByte('\n')
	b.WriteString(moduleLine("example.com/lib/sub", "sub"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(graphFooter(2, 3, 1))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))
}
```
