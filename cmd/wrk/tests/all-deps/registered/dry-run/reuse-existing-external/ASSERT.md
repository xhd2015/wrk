## Expected Output

```
would: wrk example.com/dep1 at ./external/mydep1-main-2026-06-30
would: wrked 1 deps
```

## Expected

- Exit code 0.
- Stdout is the two `would:` lines using the **existing** external name (not a new `-1` name).
- Pre-existing external path still exists; no new collision path created.
- Consumer `go.mod` still has **no** replace for dep1 (dry-run does not write).
- Stderr contains reuse warning referencing the existing abs path.

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
	wantStdout := fmt.Sprintf("would: wrk example.com/dep1 at %s\nwould: wrked 1 deps\n", wantRel)
	assert.Output(t, resp.Stdout, allDepsStdoutV2(wantStdout))

	assertFileExists(t, wantAbs)
	collided := filepath.Join(req.ConsumerTop, "external", fmt.Sprintf("mydep1-main-%s-1", wrkDate))
	assertFileNotExists(t, collided)

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep1", "") {
		t.Fatalf("dry-run must not write replace for example.com/dep1; got %+v", mod.Replace)
	}

	assertContains(t, resp.Stderr, "reusing")
	assertContains(t, resp.Stderr, wantAbs)
}
```
