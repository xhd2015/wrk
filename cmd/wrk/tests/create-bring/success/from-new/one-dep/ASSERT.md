## Expected

- Exit code 0.
- New default WT exists; stdout includes that path and the external under it.
- Last event `command=="create"`; `args` include `--new`, `--bring`, and the dep path.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wt := createBringDefaultWT(req)
	ext1 := createBringExternalPath(wt, "mydep1")
	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	if !createBringStdoutHasLine(resp.Stdout, wt) {
		t.Fatalf("stdout should include create path %q; got %q", wt, resp.Stdout)
	}
	assertFileExists(t, ext1)
	if !createBringStdoutHasLine(resp.Stdout, ext1) {
		t.Fatalf("stdout should include external %q; got %q", ext1, resp.Stdout)
	}

	mod, err := readCreateBringGoMod(wt)
	if err != nil {
		t.Fatalf("read new WT go.mod: %v", err)
	}
	if !createBringHasReplace(mod, createBringDep1Module, ext1) {
		t.Fatalf("new WT go.mod missing replace %s => %s", createBringDep1Module, ext1)
	}

	ev := createBringLastEvent(t, req.WrkHome)
	if ev.Command != "create" {
		t.Fatalf("event command: want %q, got %q args=%v", "create", ev.Command, ev.Args)
	}
	if !createBringArgsContain(ev.Args, "--new") {
		t.Fatalf("event args should include --new, got %v", ev.Args)
	}
	if !createBringArgsContain(ev.Args, "--bring") {
		t.Fatalf("event args should include --bring, got %v", ev.Args)
	}
	if !createBringArgsContain(ev.Args, req.DepPath) {
		t.Fatalf("event args should include dep %q, got %v", req.DepPath, ev.Args)
	}
}
```
