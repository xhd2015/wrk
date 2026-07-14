## Expected

- Exit code 0.
- Stdout lists only dep2, then `wrked 1 deps`.
- dep1's pre-existing replace is unchanged; dep2 is linked.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	wantDep2 := allDepsExternalAbsPath(req.ConsumerTop, "mydep2")
	wantStdout := fmt.Sprintf("wrk example.com/dep2 at %s\nwrked 1 deps\n", allDepsExternalRelPath("mydep2"))
	assert.Output(t, resp.Stdout, allDepsStdoutV2(wantStdout))

	assertFileExists(t, wantDep2)
	assertGitFileIsWorktreeLink(t, wantDep2)
	assertWorktreeListContains(t, allDepsDepMainRepo(dep2), wantDep2)

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep2", wantDep2) {
		t.Fatalf("go.mod missing replace example.com/dep2 => %s: %+v", wantDep2, mod.Replace)
	}
	dep1Replace := allDepsReplacePathForModule(mod, "example.com/dep1")
	if dep1Replace != "./external/preexisting" {
		t.Fatalf("dep1 replace should be unchanged ./external/preexisting, got %q", dep1Replace)
	}
}
```