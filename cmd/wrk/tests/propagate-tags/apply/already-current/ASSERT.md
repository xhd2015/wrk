
## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)

updated 0 modules across 0 projects
```

## Expected

- Exit code 0 (apply no-op succeeds; not "apply is not implemented").
- Source release block is printed.
- Stdout has **no** `updated example.com/app` module block.
- Footer is zero modules / zero projects with `updated` (not `would: update`).
- No `would:` prefix anywhere on stdout.

## Side Effects

- go.mod / tags / HEAD unchanged on lib and app (byte-identical go.mod).

## Exit Code

- 0

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "would:") {
		t.Fatalf("apply stdout must not use would: prefix, got %q", resp.Stdout)
	}

	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine("example.com/lib", "v1.2.3", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(applyFooter(0, 0))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))

	if strings.Contains(resp.Stdout, "updated example.com/app") {
		t.Fatalf("already-current consumer must not have updated module block, got %q", resp.Stdout)
	}

	// Full no-mutation: same as dry-run (including app go.mod bytes).
	assertDryRunNoMutation(t, req)
}
```
