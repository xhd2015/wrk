---
label: e2e, tty
explanation: requires `script` fake TTY; platform-specific
---

## Expected

- Exit code 0.
- Combined stdout+stderr does **not** contain `Proceed?`.
- Worktree removed after successful merge-back.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	combined := resp.Stdout + resp.Stderr
	assertNotContains(t, combined, "Proceed?")
	assertFileNotExists(t, req.WtDir)
}
```
