## Expected

- Non-zero exit.
- Stdout empty.
- Stderr contains refuse wording: basename, existing path, non-interactive / TTY, and that default is skip.
- Suggested shape: `wrk: myrepo already exists in <absPath>; refusing non-interactive create (default is skip; re-run in a TTY)`
- No new worktree under `{WorkRoot}/target/`.
- Prior worktree unchanged.

## Exit Code

- non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit on non-TTY named bring with existing linked WT; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout must be empty, got %q", resp.Stdout)
	}

	assertContains(t, resp.Stderr, "already exists")
	assertContains(t, resp.Stderr, req.WtDir)
	assertContains(t, resp.Stderr, "myrepo")
	// Non-interactive refuse (stable substrings from design).
	if !strings.Contains(resp.Stderr, "non-interactive") && !strings.Contains(resp.Stderr, "TTY") && !strings.Contains(resp.Stderr, "tty") {
		t.Fatalf("stderr should mention non-interactive/TTY refuse; got %q", resp.Stderr)
	}
	assertContains(t, resp.Stderr, "skip")

	assertFileExists(t, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate))
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1"))
}
```
