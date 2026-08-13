# Scenario

**Feature**: a hard post-land failure stops later peel work and reinstall tail

```
dep origin points at a non-git path -> push is fatal (not no-origin soft-skip)
-> no reinstall tail; completed land is kept
```

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T,d *session.Doctest,req *Request) error {
	_=d
	seedLinkedDep(t,req)
	// Keep origin present so this is not isNoPushRemoteErr (soft skip).
	bad := filepath.Join(req.WorkRoot, "not-a-git-remote")
	if err := os.WriteFile(bad, []byte("not a git repo\n"), 0o644); err != nil {
		return err
	}
	git(t, req.DepMain, "remote", "set-url", "origin", bad)
	req.Args=unwindGenCommitArgs(t,req,"--merge-back","--sync","--tag-next","--push","--reinstall-local")
	return nil
}
```
