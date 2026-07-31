## Expected

- Exit 0; stderr has a `warning:` from the skipped divergent sibling.

## Exit Code

- 0

```go
import (
 "strings"
 "github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T,d *session.Doctest,req *Request,resp *Response,err error) { _=d; if err!=nil {t.Fatal(err)}; assertExit0(t,resp); if !strings.Contains(strings.ToLower(resp.Stderr),"warning:") {t.Fatalf("sync warning missing from stderr: %q",resp.Stderr)} }
```
