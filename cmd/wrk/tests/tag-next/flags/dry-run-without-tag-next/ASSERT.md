## Expected

- Non-zero exit code.
- Stderr names `--dry-run` and states it is only valid with known modes (must include
  at least `--tag-next` and `--push` among the host list; full list may grow).
- Stdout empty.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	// Locked substring was outdated (missed --reinstall-local/--gen-commit-msg/--push).
	// Soft contract: error names the flag and the only-valid-with pattern.
	assertContains(t, resp.Stderr, "--dry-run is only valid with")
	assertContains(t, resp.Stderr, "--tag-next")
	assertContains(t, resp.Stderr, "--push")
}
```
