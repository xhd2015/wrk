
## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)

updated example.com/app  (project app)
  example.com/lib  v1.0.0 -> v1.2.3

updated 1 module across 1 project
```

## Expected

- Exit code **0** (partial success; build failure is soft).
- Stdout has the `updated` block and footer; **no** `go build ./... ok`;
  **no** `committed` line; **no** `would:` prefix.
- Stderr contains `warning:` and mentions build failure (exact compiler noise may vary).

## Side Effects

- App HEAD **unchanged** (no commit).
- App `go.mod` require for `example.com/lib` is bumped to `v1.2.3` (dirty tree OK).
- Source go.mod / tags / HEAD unchanged; app tags unchanged.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if strings.Contains(resp.Stdout, "would:") {
		t.Fatalf("apply stdout must not use would: prefix, got %q", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "go build ./... ok") {
		t.Fatalf("build-fail must not print go build ok, got %q", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "committed ") {
		t.Fatalf("build-fail must not print committed line, got %q", resp.Stdout)
	}

	if !strings.Contains(resp.Stderr, "warning:") {
		t.Fatalf("stderr must contain warning: prefix, got %q", resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "build") {
		t.Fatalf("stderr warning must mention build failure, got %q", resp.Stderr)
	}

	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine("example.com/lib", "v1.2.3", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(updatedHeader("example.com/app", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteString(versionBumpLine("example.com/lib", "v1.0.0", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(applyFooter(1, 1))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))

	gotMod := readFile(t, goModPath(req.AppPath))
	assertGoModRequireVersion(t, gotMod, "example.com/lib", "v1.2.3")

	assertApplyNoGitMutation(t, req)
}
```
