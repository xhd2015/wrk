## Expected

- Exit 0 with a newline-terminated expanded plan for the dependency peel.
- The plan orders generated commit, merge-back, sync, tag, push, pin, then tail reinstall.
- No ref or worktree mutation occurs.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; assertExit0(t,resp); if !strings.HasSuffix(resp.Stdout,"\n") {t.Fatalf("stdout lacks final newline: %q",resp.Stdout)}; assertContainsInOrder(t,resp.Stdout,"would: peel dep","generate","commit","merge","sync","tag","push","pin","reinstall"); if got:=git(t,req.DepMain,"rev-parse","HEAD");got!=req.BeforeDep {t.Fatalf("dry-run mutated dep: %s != %s",got,req.BeforeDep)}; if got:=git(t,req.MainRepo,"rev-parse","HEAD");got!=req.BeforeMain {t.Fatalf("dry-run mutated main: %s != %s",got,req.BeforeMain)} }
```
