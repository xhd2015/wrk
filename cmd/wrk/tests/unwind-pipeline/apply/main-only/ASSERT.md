## Expected

- Exit 0. A main checkout must not invoke generated commit or merge-back.

## Exit Code

- 0

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; assertExit0(t,resp); out:=strings.ToLower(resp.Stdout+resp.Stderr); if strings.Contains(out,"generate")||strings.Contains(out,"merge-back") {t.Fatalf("main-only peel must skip generated commit and merge-back: %q",out)} }
```
