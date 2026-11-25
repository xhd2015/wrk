## Expected

- Non-zero exit.
- Stderr names `--new` and requires a `<project>` argument.
- No new worktree under `{WRK_HOME}/worktrees`.
- No `src/external/`; no new branch in `src`.

## Errors

- `--new` with `--bring` requires an explicit `<project>` positional.

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
		t.Fatalf("expected non-zero for --new --bring without project, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--new") {
		t.Fatalf("stderr should name --new, got %q", se)
	}
	if !strings.Contains(se, "<project>") {
		t.Fatalf("stderr should require a <project> argument, got %q", se)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("rejection must not print a create path; stdout=%q", resp.Stdout)
	}
	assertFileNotExists(t, createBringDefaultWT(req))
	assertFileNotExists(t, filepath.Join(req.MainRepo, "external"))
	assertBranchNotExists(t, req.MainRepo, branchName("master", wrkDate, 0))
	if names := createBringListHomeWTs(t, req); len(names) != 0 {
		t.Fatalf("rejection must not create worktrees; found %v", names)
	}
}
```
