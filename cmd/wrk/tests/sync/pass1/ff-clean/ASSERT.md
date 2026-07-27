---
label: slow
explanation: isolated git init + worktree add + two commits; apply FF merge
---

## Expected Output

```
main ← feature-login  (+2 commits)

synced: 1 into main, 0 into worktrees, 0 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly the detail line, blank line, and summary above (trailing `\n`).
- Stderr empty.
- Side effect: `git rev-parse HEAD` on main equals pre-run worktree tip (`req.WtSHA`).
- Worktree tip unchanged (`req.WtSHA`).

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
		1, 0, 0, false,
	)
	assertOutputExact(t, resp.Stdout, syncStdoutV2(want))

	assertHEAD(t, req.MainRepo, req.WtSHA)
	assertHEAD(t, req.WtPath, req.WtSHA)
}
```
