## Expected

- Exit code 0.
- Stdout is embedded `SKILL.md` (marker + `name: wrk`).
- Stdout documents **`--new`** as create (or `wrk --new` examples).
- Stdout documents **dashboard** and/or that bare `wrk` opens the dashboard / does **not** create a worktree by default.
- Must **not** still claim bare `wrk` alone is “create from cwd” without mentioning the dashboard/`--new` split.
- Stderr empty.

## Side Effects

- Read-only.

## Exit Code

- 0

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmbeddedSkillStdoutDash(t, resp.Stdout)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	out := resp.Stdout
	lower := strings.ToLower(out)
	if !strings.Contains(out, "--new") && !strings.Contains(out, "wrk --new") {
		t.Fatalf("embedded SKILL.md should document --new create entry; stdout:\n%s", out)
	}
	hasDashboard := strings.Contains(lower, "dashboard")
	hasNoBareCreate := strings.Contains(lower, "does not create") ||
		strings.Contains(lower, "no create") ||
		strings.Contains(lower, "not create") ||
		(strings.Contains(lower, "bare") && strings.Contains(lower, "dashboard"))
	if !hasDashboard && !hasNoBareCreate {
		t.Fatalf("embedded SKILL.md should document dashboard / bare no-create; stdout:\n%s", out)
	}
	// Guard against stale “wrk # create from cwd” as the only create story without --new context.
	if strings.Contains(out, "wrk                              # create from cwd") &&
		!strings.Contains(out, "--new") {
		t.Fatalf("SKILL still documents bare wrk as create without --new; stdout:\n%s", out)
	}
}
```
