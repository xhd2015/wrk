---
label: slow
explanation: isolated git + dirty main skip
---

## Expected Output

```
synced: 0 into main, 0 into worktrees, 1 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly zero-action summary with skipped=1.
- Stderr contains `warning: skip feature-login: dirty main`.
- Main HEAD tip unchanged (dirty tree does not move branch tip).
- Worktree HEAD unchanged.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantOut := buildSyncStdout(nil, 0, 0, 1, false)
	assertOutputExact(t, resp.Stdout, syncStdoutV2(wantOut))
	assertContains(t, resp.Stderr, "warning: skip feature-login: dirty main")

	assertHEADUnchanged(t, req.MainRepo, req.MainSHA)
	assertHEADUnchanged(t, req.WtPath, req.WtSHA)
}
```
