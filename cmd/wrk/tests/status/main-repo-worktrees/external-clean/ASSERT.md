## Expected

- Exit code 0.
- Scan block for main with invocation-cwd Dir (`.` from main root) and `Remote:`.
- Appended full block for external wt with `statusDirLine` Dir (typically relative
  `../.wrk/worktrees/…` for `{WorkRoot}/.wrk` fixtures — **not** forced absolute) and
  `Master: identical`.
- Blank line between blocks.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 2)

	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
		appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
	))
}
```
