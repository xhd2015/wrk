## Expected

- Non-zero exit from push (there is intentionally no remote).
- The tail reinstall stage is absent; completed land work is not rolled back.

## Exit Code

- non-zero

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; if resp.ExitCode==0 {t.Fatalf("fatal push unexpectedly succeeded: stdout=%q stderr=%q",resp.Stdout,resp.Stderr)}; if strings.Contains(strings.ToLower(resp.Stdout+resp.Stderr),"reinstall") {t.Fatalf("tail reinstall ran after fatal stage: stdout=%q stderr=%q",resp.Stdout,resp.Stderr)}; if got:=git(t,req.DepMain,"rev-parse","HEAD");got==req.BeforeDep {t.Fatal("completed land should not be rolled back") } }
```
