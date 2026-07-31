## Expected

- Exit 0; help lists `--unwind`, `--sync`, and `--gen-commit-msg`.

## Exit Code

- 0

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; assertExit0(t,resp); for _,flag:=range []string{"--unwind","--sync","--gen-commit-msg"} { if !strings.Contains(resp.Stdout,flag) {t.Fatalf("help missing %s: %q",flag,resp.Stdout)} } }
```
