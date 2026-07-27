# Scenario

**Feature**: wrapper create with explicit `<target-dir>` does not auto-cd even from home

```
# StartDir = FakeHome; auto-cd channel open (default wrapper)
# second positional <target-dir> missing; parent exists
source bash.sh; wrk <mainRepo> <target>
  -> stdout worktree path; no stderr cd; FinalPWD stays FakeHome
```

## Steps

1. Init main repo; start shell cwd at FakeHome (user home).
2. Choose absolute `<target>` under WorkRoot that does not exist (parent exists).
3. Invoke `wrk <mainRepo> <target>` via installed wrapper (default auto-cd on).
4. Expect create success, no stderr follow-up `cd`, shell stays at FakeHome.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := setupMainRepo(t, req)
	target := filepath.Join(req.WorkRoot, "wt")
	req.RepoDir = mainRepo
	// Shell StartDir = FakeHome so home gate would open for default create.
	req.StartDir = req.FakeHome
	req.CLIArgs = []string{mainRepo, target}
	return nil
}
```
