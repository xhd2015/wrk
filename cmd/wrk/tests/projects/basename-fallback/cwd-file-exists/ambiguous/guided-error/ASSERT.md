## Expected

- Non-zero exit code.
- Stderr reports cwd file path `exists and is a file`.
- Stderr lists both registered projects (lexicographically sorted absolute paths).
- Stderr hint suggests `wrk <full-path> --status` (literal `<full-path>` placeholder).
- Stdout is empty.
- No worktree or status output produced.

## Errors

- Ambiguous basename with cwd file collision; user must pick a full project path.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}

	filePath := resolvePath(t, filepath.Join(req.RepoDir, "spl"))
	sorted := sortedSavedPaths(t, req.MainRepo, req.SecondRepo)

	assert.Output(t, resp.Stderr, `<contains>
`+filePath+` exists and is a file
</contains>`)

	assert.Output(t, resp.Stderr, `<contains>
"spl" matches registered project(s):
  `+sorted[0]+`
  `+sorted[1]+`
</contains>`)

	// assert.Output cannot embed literal <full-path> — doctest treats it as a tag.
	wantHint := "use `wrk <full-path> --status` instead"
	if !strings.Contains(resp.Stderr, wantHint) {
		t.Fatalf("stderr missing hint %q, got:\n%s", wantHint, resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, "spl", "main", wrkDate, 0)
	assertFileNotExists(t, wantPath)
	assertWorktreeListNotContains(t, req.MainRepo, wantPath)
	assertWorktreeListNotContains(t, req.SecondRepo, wantPath)
}
```