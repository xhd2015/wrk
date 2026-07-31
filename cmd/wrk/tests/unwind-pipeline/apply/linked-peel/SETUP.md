# Scenario

**Feature**: linked peel generates its commit before land and release stages

```
dirty linked dependency -> generated commit -> merge-back -> sync -> tag -> push
```

```go
import "github.com/xhd2015/doctest/session"
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; seedLinkedDep(t,req); req.Args=unwindGenCommitArgs(t,req,"--merge-back","--sync","--tag-next","--push"); return nil }
```
