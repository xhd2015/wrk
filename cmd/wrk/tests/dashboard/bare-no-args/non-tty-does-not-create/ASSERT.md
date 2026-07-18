## Expected

- **Critical**: no new worktree under `{WRK_HOME}/worktrees/` (create must not run).
- Exit code may be 0 (dashboard stub success) or non-zero (stub error / not implemented).
- Do **not** require full dashboard snapshot chrome (`[•]` glyphs, stage list) in P1.
- Stdout must **not** be a newly created worktree path that exists on disk under worktrees.

## Side Effects

- None from create mode (no linked worktree, no branch `main-2026-06-30` for this invocation).

## Exit Code

- Any (0 or non-zero); create absence is the gate.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertNoWorktreesCreated(t, req)

	// If implementation still runs old bare-create, stdout is the wt path and
	// assertNoWorktreesCreated already failed. Extra guard: stdout path must not
	// exist as a create worktree when it looks like WRK_HOME/worktrees/*.
	out := strings.TrimSpace(resp.Stdout)
	if out != "" {
		want := wantDashboardCreateWorktree(req)
		if out == want {
			t.Fatalf("bare no-args still looks like create: stdout=%q", resp.Stdout)
		}
		if strings.HasPrefix(out, worktreesRoot(req)+string(filepath.Separator)) {
			if _, statErr := os.Stat(out); statErr == nil {
				t.Fatalf("bare no-args must not print an existing create worktree path: %q", out)
			}
		}
	}
}
```
