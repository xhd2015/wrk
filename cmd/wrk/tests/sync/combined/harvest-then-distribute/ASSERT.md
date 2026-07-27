---
label: slow
explanation: two linked worktrees; harvest then distribute apply
---

## Expected Output

```
main ← feature-login  (+2 commits)
feature-api ← main  (+2 commits)

synced: 1 into main, 1 into worktrees, 0 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly the two detail lines (pass1 then pass2), blank line, summary.
- Stderr empty.
- Side effects after apply:
  - main HEAD == feature-login tip (`req.WtSHA`)
  - feature-api HEAD == main tip (`req.WtSHA`)
  - feature-login tip still `req.WtSHA`

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
		[]string{
			syncDetailPass1("feature-login", 2),
			syncDetailPass2("feature-api", 2),
		},
		1, 1, 0, false,
	)
	assertOutputExact(t, resp.Stdout, syncStdoutV2(want))

	// After harvest, main and login share login tip; pass2 FF api to that tip.
	assertHEAD(t, req.MainRepo, req.WtSHA)
	assertHEAD(t, req.WtPath, req.WtSHA)
	assertHEAD(t, req.Wt2Path, req.WtSHA)
}
```
