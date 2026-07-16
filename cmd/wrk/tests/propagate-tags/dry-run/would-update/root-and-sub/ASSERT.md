## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)
  example.com/lib/sub  @ v0.1.0  (tag sub/v0.1.0)

would: update example.com/app  (project app)
  example.com/lib  v1.0.0 -> v1.2.3
  example.com/lib/sub  v0.0.1 -> v0.1.0

would: update 2 modules across 1 project
```

## Expected

- Exit code 0.
- Stdout matches the plan shape above (source path is absolute fixture path).
- Source block lists root and nested sub releases.
- One `would: update` block for `example.com/app` with both version arrows.
- Footer `would: update 2 modules across 1 project`.
- Stderr empty (or only soft warnings unrelated to this fixture — expect empty).

## Side Effects

- Consumer and source `go.mod` bytes unchanged.
- No git tag changes on lib or app.
- HEAD unchanged on lib and app.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}

	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine("example.com/lib", "v1.2.3", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine("example.com/lib/sub", "v0.1.0", "sub/v0.1.0"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(wouldUpdateHeader("example.com/app", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteString(versionBumpLine("example.com/lib", "v1.0.0", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteString(versionBumpLine("example.com/lib/sub", "v0.0.1", "v0.1.0"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(planFooter(2, 1))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))

	assertDryRunNoMutation(t, req)
}
```
