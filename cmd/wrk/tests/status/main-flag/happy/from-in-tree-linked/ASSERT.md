## Expected

- Exit code 0; stderr empty.
- Stdout equals `wrk --status` from main (full main-repo status: root `.` with `Remote:`, plus relative in-tree linked block with `Master:`).
- Stdout is **not** equal to plain `wrk --status` from the in-tree linked cwd (linked-cwd shortcut: single `.` + `Master` only).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutEqualsMainStatus(t, req, resp)

	// Prove we did not take the linked-in-tree cwd special path.
	plainLinked := runWrkCapture(t, req, req.WtDir, "--status")
	if plainLinked.ExitCode != 0 {
		t.Fatalf("plain --status from in-tree linked exit %d stderr=%q", plainLinked.ExitCode, plainLinked.Stderr)
	}
	if resp.Stdout == plainLinked.Stdout {
		t.Fatalf("--main --status must not match linked-cwd shortcut status:\n%s", resp.Stdout)
	}
}
```