## Expected

- Exit code 0.
- Stdout: `wrk example.com/dep1 at ./external/mydep1-main-{date}` then `wrked 1 deps`.
- External path is the pre-existing one (no `…-1`).
- Exactly one entry under `external/`.
- go.mod replace for dep1 points at the reused abs path.
- Stderr contains reuse warning for the existing external path.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantAbs := req.ExternalWtDir
	wantRel := allDepsExternalRelPath("mydep1")
	wantStdout := fmt.Sprintf("wrk example.com/dep1 at %s\nwrked 1 deps\n", wantRel)
	assert.Output(t, resp.Stdout, allDepsStdoutV2(wantStdout))

	assertFileExists(t, wantAbs)
	assertGitFileIsWorktreeLink(t, wantAbs)
	assertWorktreeListContains(t, allDepsDepMainRepo(req.DepPath), wantAbs)
	assertFileNotExists(t, allDepsExternalAbsPath(req.ConsumerTop, "mydep1")+"-1")
	// Also ensure no collision-named sibling under external/.
	collided := filepath.Join(req.ConsumerTop, "external", fmt.Sprintf("mydep1-main-%s-1", wrkDate))
	assertFileNotExists(t, collided)

	entries, err := os.ReadDir(filepath.Join(req.ConsumerTop, "external"))
	if err != nil {
		t.Fatalf("readdir external: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 external entry, got %d", len(entries))
	}

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep1", wantAbs) {
		t.Fatalf("go.mod missing replace example.com/dep1 => %s: %+v", wantAbs, mod.Replace)
	}

	assertContains(t, resp.Stderr, "already exists under external/")
	assertContains(t, resp.Stderr, "reusing")
	assertContains(t, resp.Stderr, wantAbs)
}
```
