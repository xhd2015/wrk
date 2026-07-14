## Expected Output

```
would: wrk example.com/dep1 at ./external/mydep1-main-2026-06-30
would: wrk example.com/dep2 at ./external/mydep2-main-2026-06-30
would: wrked 2 deps
```

## Expected

- Exit code 0.
- Stdout is exactly the three `would:` lines above (project path order: mydep1 before mydep2).
- `{consumerTop}/external/` does NOT exist.
- Consumer `go.mod` has NO new replaces.
- Consumer `.gitignore` has NO `/external` line.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantStdout := fmt.Sprintf("would: wrk example.com/dep1 at %s\nwould: wrk example.com/dep2 at %s\nwould: wrked 2 deps\n",
		allDepsExternalRelPath("mydep1"), allDepsExternalRelPath("mydep2"))
	assert.Output(t, resp.Stdout, allDepsStdoutV2(wantStdout))

	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
	wantDep1 := allDepsExternalAbsPath(req.ConsumerTop, "mydep1")
	wantDep2 := allDepsExternalAbsPath(req.ConsumerTop, "mydep2")
	assertFileNotExists(t, wantDep1)
	assertFileNotExists(t, wantDep2)

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep1", "") {
		t.Fatalf("go.mod should have no replace for example.com/dep1 after dry-run")
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep2", "") {
		t.Fatalf("go.mod should have no replace for example.com/dep2 after dry-run")
	}

	hasExternal, err := allDepsGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if hasExternal {
		t.Fatalf(".gitignore should have no /external line after dry-run")
	}
}
```