## Expected

- Non-zero exit code.
- Stderr contains substring `--dry-run is only valid with --all-deps` (existing contract).
- Stderr also mentions `--sync` once the dry-run allowlist includes sync.
- Stdout empty.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	// Keep existing all-deps substring so older dry-run tests stay green.
	assertContains(t, resp.Stderr, "--dry-run is only valid with --all-deps")
	// Phase 1: message should mention --sync when allowlist is updated.
	assertContains(t, resp.Stderr, "--sync")
}
```
