## Expected

- Non-zero exit.
- Stderr is a **bring resolve** failure (does not exist / no match / unresolvable), **not** create/`--bring` mutual exclusion.
- No new worktree under `{WRK_HOME}/worktrees`.
- Expected create path `src-main-2026-06-30` does not exist.

## Errors

- Preflight cannot resolve the bring basename `no-such-basename`.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for unresolvable bring basename, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if strings.Contains(se, "mutually exclusive") || strings.Contains(se, "exclusive") {
		t.Fatalf("want bring preflight resolve error, not create/--bring exclusivity; stderr=%q", resp.Stderr)
	}
	if !strings.Contains(se, "does not exist") &&
		!strings.Contains(se, "no-such-basename") &&
		!strings.Contains(se, "not found") &&
		!strings.Contains(se, "unresolv") {
		t.Fatalf("stderr should name the unresolvable bring path; got %q", resp.Stderr)
	}
	want := createBringDefaultWT(req)
	assertFileNotExists(t, want)
	if names := createBringListHomeWTs(t, req); len(names) != 0 {
		t.Fatalf("preflight fail must not create worktrees; found %v", names)
	}
}
```
