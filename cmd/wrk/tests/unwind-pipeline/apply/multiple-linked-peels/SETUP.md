# Scenario

**Feature**: commit generation is scoped to each eligible linked peel

```
two dirty linked dependencies -> unwind -> one generated commit for each peel
```

```go
import (
 "os"
 "path/filepath"
 "github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; seedLinkedDep(t,req); second:=filepath.Join(req.WorkRoot,"dep-two"); if err:=os.MkdirAll(second,0o755);err!=nil{return err}; git(t,second,"init","-b","main"); git(t,second,"config","core.hooksPath","/dev/null"); configureRepoGitIdent(t,second); if err:=os.WriteFile(filepath.Join(second,"go.mod"),[]byte("module example.com/dep-two\n\ngo 1.22\n"),0o644);err!=nil{return err}; git(t,second,"add","-f","go.mod"); git(t,second,"commit","-m","initial"); remote:=filepath.Join(req.WorkRoot,"dep-two-origin.git"); git(t,req.WorkRoot,"init","--bare",remote); setOrigin(t,second,remote); wt:=filepath.Join(req.MainRepo,"external","dep-two"); git(t,second,"worktree","add","-b","feature-two",wt); if err:=os.WriteFile(filepath.Join(wt,"change.txt"),[]byte("second\n"),0o644);err!=nil{return err}; req.SiblingWorktree=wt; req.Args=unwindGenCommitArgs(t,req,"--merge-back","--sync","--tag-next","--push"); return nil }
```
