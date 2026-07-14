## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WorkRoot}/wt` (fixed spawn path).
- Branch checked out is `main-2026-06-30-1` (new branch; preferred name was taken).
- Stderr contains timestamp `worktree add` pre-command log line including `-b` (new branch).
- Stderr does **not** contain `--no-checkout` (branch-reuse path removed).
- Stderr contains git `worktree add` subprocess output (`Preparing worktree` or `HEAD is now at`).

## Exit Code

- 0

```go
import "path/filepath"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := filepath.Join(req.WorkRoot, "wt")
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	wantBranch := branchName("main", wrkDate, 1)
	assertBranchExists(t, req.TargetDir, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

	assertStderrContainsGitSubcommand(t, resp.Stderr, "worktree add")
	assertContains(t, resp.Stderr, "-b")
	assertNotContains(t, resp.Stderr, "--no-checkout")
	assertStderrVerboseFormat(t, resp.Stderr)
	assertStderrContainsWorktreeAddOutput(t, resp.Stderr)
}
```
