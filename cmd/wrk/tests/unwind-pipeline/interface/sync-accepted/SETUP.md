# Scenario

**Feature**: sync is an unwind partner

```
dirty main -> --unwind --sync --dry-run -> accepted plan
```

```go
import (
 "os"
 "path/filepath"
 "github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; os.WriteFile(filepath.Join(req.MainRepo,"dirty"),[]byte("x\n"),0o644); req.Args=[]string{"--unwind","--sync","--dry-run"}; return nil }
```
