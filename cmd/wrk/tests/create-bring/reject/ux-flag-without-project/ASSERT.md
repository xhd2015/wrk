## Expected

- Non-zero exit.
- Stderr names `--no-new-terminal` and requires a `<project>` argument.
- Stderr is **not** a create/agent failure (no worktree path on stdout).
- No new worktree under `{WRK_HOME}/worktrees`.
- No `src/external/`; no new branch in `src`.

## Errors

- Create selectors with `--bring` require an explicit `<project>` positional.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --bring --no-new-terminal without project, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--no-new-terminal") {
		t.Fatalf("stderr should name --no-new-terminal, got %q", se)
	}
	if !strings.Contains(se, "<project>") {
		t.Fatalf("stderr should require a <project> argument, got %q", se)
	}
	if strings.Contains(se, "agent-run") {
		t.Fatalf("rejection must happen before agent launch, got %q", se)
	}
	// Zero side effects: no create path printed, no WT, no external, no branch.
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("rejection must not print a create path; stdout=%q", resp.Stdout)
	}
	assertFileNotExists(t, createBringDefaultWT(req))
	assertFileNotExists(t, filepath.Join(req.MainRepo, "external"))
	assertBranchNotExists(t, req.MainRepo, branchName("main", wrkDate, 0))
	if names := createBringListHomeWTs(t, req); len(names) != 0 {
		t.Fatalf("rejection must not create worktrees; found %v", names)
	}
}
```
