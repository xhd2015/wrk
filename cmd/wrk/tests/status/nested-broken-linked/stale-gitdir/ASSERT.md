## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <root hash>  root repo
Status:       clean
Remote:       (no upstream)

Dir:          tools/good
Branch:       main
Commit:       <good hash>  good child
Status:       clean

Dir:          vendor/host
Branch:       main
Commit:       <host hash>  host repo
Status:       clean

Dir:          vendor/host/broken-wt
Status:       error: fatal: not a git repository: <stale gitdir path>
```

## Expected

- Exit code 0 (broken worktree does not abort the run).
- Stdout has four blocks in `scan_repo` path order; healthy blocks are full; broken block is minimal (`Dir` + `Status: error: …` only).
- Broken block `Dir` is **relative** (`vendor/host/broken-wt`), not absolute.
- Stderr is empty.

## Side Effects

- No repository files are changed.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 4)

	errLine := scanErrorStatusPlain(t, req.WtDir)
	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		statusRootBlockPlain(t, req.MainRepo, "clean", statusNoUpstreamRemote()),
		statusBlockPlain(t, req.DepPath, "tools/good", "clean"),
		statusBlockPlain(t, req.ConsumerTop, "vendor/host", "clean"),
		scanBrokenBlockPlain("vendor/host/broken-wt", errLine),
	))
}
```