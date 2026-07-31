## Expected

- Non-zero validation error naming the incompatible unwind mode.

## Exit Code

- non-zero

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; if resp.ExitCode==0 {t.Fatal("unrelated --list must remain incompatible")}; got:=resp.Stdout+resp.Stderr; if !strings.Contains(got,"--unwind") {t.Fatalf("incompatibility should name unwind: %q",got)} }
```
