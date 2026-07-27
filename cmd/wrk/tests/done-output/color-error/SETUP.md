# Scenario

**Feature**: color `Error:` red (and `warning:` yellow) when color policy on (D9)

```
# --color + cascade/preflight hard error → stderr Error: prefix uses red CSI
# pipe default / no --color → no ANSI on Error:
```

## Preconditions

- Color on: `--color` forces ANSI even on doctest pipes.
- Prefer go-best-practice cli/color semantics; prefix-only coloring is enough.
- Classic TDD: fatal cascade/preflight errors not colored today → force-color leaf **RED**.

## Steps

- Grouping: leaves build nested-main or dirty preflight fixtures and set Args.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}

// assertStderrHasRedErrorPrefix requires red CSI coloring the Error: prefix
// (prefix-only color is the preferred product shape).
func assertStderrHasRedErrorPrefix(t *testing.T, stderr string) {
	t.Helper()
	if !strings.Contains(stderr, "Error:") {
		t.Fatalf("expected Error: in stderr; got:\n%s", stderr)
	}
	// Preferred shapes: colorize("Error:", red) or red CSI immediately before Error:.
	if strings.Contains(stderr, "\x1b[31mError:") || strings.Contains(stderr, "\033[31mError:") {
		return
	}
	// Fallback: red CSI appears in a short window before Error: on the same line.
	for _, line := range strings.Split(stderr, "\n") {
		if !strings.Contains(line, "Error:") {
			continue
		}
		idx := strings.Index(line, "Error:")
		prefix := line[:idx]
		if strings.Contains(prefix, "\x1b[31m") || strings.Contains(prefix, "\033[31m") {
			return
		}
	}
	t.Fatalf("with --color, Error: prefix must use red ANSI (\\x1b[31m); stderr:\n%q", stderr)
}

var _ = assertStderrHasRedErrorPrefix
```
