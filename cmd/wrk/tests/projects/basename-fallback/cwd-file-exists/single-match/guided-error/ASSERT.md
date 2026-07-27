## Expected

- Non-zero exit code.
- Stderr reports cwd file path `exists and is a file`.
- Stderr lists one registered project at the saved repo absolute path.
- Stderr hint suggests `wrk <saved-path> -t 'optimize skills output'`.
- Stdout is empty.
- No worktree created under `{WRK_HOME}/worktrees/`.

## Errors

- Cwd file blocks basename resolution; user must pass the full saved project path.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"

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

	filePath := resolvePath(t, filepath.Join(req.RepoDir, "myrepo"))
	savedPath := resolvePath(t, req.MainRepo)

	assert.Output(t, resp.Stderr, `<contains>
`+filePath+` exists and is a file
</contains>`)

	assert.Output(t, resp.Stderr, `<contains>
"myrepo" matches registered project(s):
  `+savedPath+`
</contains>`)

	hint := "wrk " + savedPath + " -t 'optimize skills output'"
	assert.Output(t, resp.Stderr, `<contains>
use `+"`"+hint+"`"+` instead
</contains>`)

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertFileNotExists(t, wantPath)
	assertWorktreeListNotContains(t, req.MainRepo, wantPath)
}
```