## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree.
- Both `go-pkgs/go.mod` and `tools/go.mod` have `replace example.com/dep => <external path>`.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	// Both sub-modules should have the replace.
	for _, dir := range []string{req.ConsumerModDir, req.ConsumerModDir2} {
		mod, err := readGoMod(dir)
		if err != nil {
			t.Fatalf("read go.mod in %s: %v", dir, err)
		}
		if !hasReplaceForModule(mod, depModulePath, wantPath) {
			t.Fatalf("%s/go.mod missing replace %s => %s: %+v", dir, depModulePath, wantPath, mod.Replace)
		}
	}

	ok, err := gitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}
}
```