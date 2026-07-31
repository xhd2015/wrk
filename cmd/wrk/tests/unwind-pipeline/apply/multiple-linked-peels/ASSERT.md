## Expected

- Exit 0 and generate/commit once for each dirty linked worktree, not once globally.

## Exit Code

- 0

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; assertExit0(t,resp); if n:=strings.Count(strings.ToLower(resp.Stdout+resp.Stderr),"generate"); n<2 {t.Fatalf("want per-peel generation >=2, got %d: stdout=%q stderr=%q",n,resp.Stdout,resp.Stderr)} }
```
