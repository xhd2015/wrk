## Expected

- Exit code 0.
- `Status:` uses granular coloring: red for `dirty` and `2 changed`, grey (`#90`) for zero-count segments.
- Separators `(`, `, `, `)` are uncolored.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertColorProjectsBlocksSeparated(t, resp.Stdout, 1)
	plain := stripANSI(resp.Stdout)
	if plain == resp.Stdout {
		t.Fatalf("expected ANSI color codes in --color output, got plain:\n%s", resp.Stdout)
	}
	if !strings.Contains(plain, "Status:") || !strings.Contains(plain, "dirty") {
		t.Fatalf("expected dirty Status line, got:\n%s", plain)
	}
	if !strings.Contains(plain, "2 changed") {
		t.Fatalf("expected 2 changed in Status, got:\n%s", plain)
	}
	if !strings.Contains(plain, "Worktrees:") {
		t.Fatalf("expected Worktrees line, got:\n%s", plain)
	}
}
```