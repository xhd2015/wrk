## Expected

- Exit code 0; replace present (tidy ran silently).
- Stderr does **not** contain a `$ go … mod tidy` pre-line pattern.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, bringDepModulePath, wantPath) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", bringDepModulePath, wantPath, mod.Replace)
	}

	assertBringStderrNoTidyPreLine(t, resp.Stderr)
}
```
