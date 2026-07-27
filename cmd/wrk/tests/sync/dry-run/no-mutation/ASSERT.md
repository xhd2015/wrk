---
label: slow
explanation: dry-run FF plan; assert no ref mutation
---

## Expected Output

```
would: main ← feature-login  (+2 commits)

would: synced: 1 into main, 0 into worktrees, 0 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly the would-prefixed detail + blank line + would-summary.
- Stderr empty.
- Side effects: main HEAD still `req.MainSHA`; worktree HEAD still `req.WtSHA`
  (no `merge --ff-only`).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	want := buildSyncStdout(
		[]string{syncDetailPass1("feature-login", 2)},
		1, 0, 0, true,
	)
	assertOutputExact(t, resp.Stdout, syncStdoutV2(want))

	assertHEADUnchanged(t, req.MainRepo, req.MainSHA)
	assertHEADUnchanged(t, req.WtPath, req.WtSHA)
}
```
