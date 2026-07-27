# Scenario

**Feature**: wrk --rm deletes a recorded main repository path

```
wrk --rm <dir> -> resolve path -> delete matching projects.json entry -> stdout main path (or empty if idempotent)
```

## Preconditions

- `wrk --rm` is a standalone mode; mutually exclusive with other modes.
- Requires a non-empty path argument after `--rm`.
- Success removes the entry from `projects.json` only (no worktree/git/history deletion).

## Steps

- Descendants set `req.Args = []string{"--rm", <path>}` (or invalid combinations for error leaves).

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}

func recordProjectViaAdd(t *testing.T, req *Request, path string) {
	t.Helper()
	runWrkWithArgs(t, req, req.WorkRoot, "--add", path)
}

func assertStdoutEmpty(t *testing.T, stdout string) {
	t.Helper()
	assert.Output(t, stdout, v2StdoutTemplate(""))
}

func assertProjectNotRecorded(t *testing.T, wrkHome, path string) {
	t.Helper()
	pf := readProjectsFile(t, wrkHome)
	want := resolvePath(t, path)
	for _, p := range pf.Projects {
		got := resolvePath(t, p.Path)
		if got == want {
			t.Fatalf("projects.json should not contain %q, got %+v", want, pf.Projects)
		}
	}
}

func removeGitDir(t *testing.T, repoDir string) {
	t.Helper()
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		t.Fatalf("remove %s: %v", gitDir, err)
	}
}

func ensureRemoveHelpersUsed() {
	_ = recordProjectViaAdd
	_ = assertStdoutEmpty
	_ = assertProjectNotRecorded
	_ = removeGitDir
}
```
