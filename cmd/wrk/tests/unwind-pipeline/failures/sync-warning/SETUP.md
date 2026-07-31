# Scenario

**Feature**: sync warnings do not fail a completed peel

```
linked dependency plus stale sibling -> sync -> warning on stderr and successful unwind
```

```go
import (
 "path/filepath"
 "github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; seedLinkedDep(t,req); req.SiblingWorktree=filepath.Join(req.WorkRoot,"stale-sibling"); git(t,req.DepMain,"worktree","add","-b","stale",req.SiblingWorktree); git(t,req.SiblingWorktree,"commit","--allow-empty","-m","diverge sibling"); req.Args=unwindGenCommitArgs(t,req,"--merge-back","--sync"); return nil }
```
