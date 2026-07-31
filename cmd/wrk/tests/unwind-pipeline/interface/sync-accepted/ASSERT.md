## Expected

- `--unwind --sync` is accepted (exit 0), rather than reported mutually exclusive.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/session"
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; assertExit0(t,resp) }
```
