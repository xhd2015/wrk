## Expected

- Exit 0.
- Stdout contains `would: pin-local example.com/consumer <- example.com/dep => `
  followed by a relative path (`./` or `../`).
- Prefer `./external/dep` when cwd is primary root.
- No bare apply `pin-local` lines (must use `would:` prefix).
- go.mod identical to baseline (no writes).

## Side Effects

- None: dry-run must not mutate go.mod.

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
	assertWouldPinLocalLine(t, resp.Stdout, modConsumer, modDep)
	foundRel := false
	for _, line := range strings.Split(resp.Stdout, "\n") {
		trim := strings.TrimSpace(line)
		prefix := "would: pin-local " + modConsumer + " <- " + modDep + " => "
		if !strings.HasPrefix(trim, prefix) {
			continue
		}
		rel := strings.TrimSpace(strings.TrimPrefix(trim, prefix))
		if strings.HasPrefix(rel, "./") || strings.HasPrefix(rel, "../") {
			foundRel = true
			if !strings.Contains(rel, "external") {
				t.Fatalf("expected relative path under external/, got %q", rel)
			}
			if filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") {
				t.Fatalf("dry-run path must be relative, got %q", rel)
			}
		}
	}
	if !foundRel {
		t.Fatalf("would: pin-local line must use relative NewPath; stdout=%q", resp.Stdout)
	}
	for _, line := range strings.Split(resp.Stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "pin-local ") {
			t.Fatalf("dry-run must not emit bare pin-local lines: %q", trim)
		}
	}
	assertGoModUnchanged(t, req)
}
```
