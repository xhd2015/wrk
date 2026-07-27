# Scenario

**Feature**: <target-dir> + --list → wrk: unexpected arguments (target-dir is create-only)

```
# --list is not a create path; a second positional is rejected
myrepo (main) -> wrk myrepo <target-dir> --list -> non-zero, wrk: unexpected arguments
```

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Set `req.SpawnDir = {WorkRoot}/wt`.
3. Set `req.Args = ["--list"]`.
4. Run `wrk myrepo {WorkRoot}/wt --list` from process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"--list"}
	return nil
}
```
