## Expected

- Non-zero exit code.
- Stderr contains the locked dry-run host list including `--propagate-tags`:
  `--dry-run is only valid with --done, --merge-back, --all-deps, --tag-next, --propagate-tags, or --sync`
- Stdout empty.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "--dry-run is only valid with --done, --merge-back, --all-deps, --tag-next, --propagate-tags, or --sync")
}
```
