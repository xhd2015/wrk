# Scenario

**Feature**: a main-only pending member has no generated commit or land stage

```
dirty app main -> unwind with all modifiers -> sync/release only
```

```go
import (
 "os"
 "path/filepath"
 "github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; seedMain(t,req); os.WriteFile(filepath.Join(req.MainRepo,"dirty.txt"),[]byte("dirty\n"),0o644); req.BeforeMain=git(t,req.MainRepo,"rev-parse","HEAD"); req.Args=unwindGenCommitArgs(t,req,"--merge-back","--sync","--tag-next","--push"); return nil }
```
