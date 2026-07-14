## Expected

- Exit code 0.
- Fake shell launched at main root (nested shell path).
- Follow-up file remains empty (no in-place `cd` write).
- Minimal launch UX (empty stdout; no install hint).

## Side Effects

- Shell only; WRK_FOLLOWUP_FILE content unchanged (empty).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertMinimalLaunchUX(t, resp)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, req.MainRepo)
	assertFollowupEmpty(t, req)
}
```
