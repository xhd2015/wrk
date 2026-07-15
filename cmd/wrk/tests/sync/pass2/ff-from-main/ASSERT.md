---
label: slow
explanation: isolated git + pass2 FF apply
---

## Expected Output

```
feature-login ← main  (+1 commit)

synced: 0 into main, 1 into worktrees, 0 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly the pass-2 detail line, blank line, and summary.
- Stderr empty.
- Side effect: worktree HEAD equals main tip (`req.MainSHA`).
- Main tip unchanged.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	want := buildSyncStdout(
		[]string{syncDetailPass2("feature-login", 1)},
		0, 1, 0, false,
	)
	assertOutputExact(t, resp.Stdout, syncStdoutV2(want))

	assertHEAD(t, req.MainRepo, req.MainSHA)
	assertHEAD(t, req.WtPath, req.MainSHA)
}
```
