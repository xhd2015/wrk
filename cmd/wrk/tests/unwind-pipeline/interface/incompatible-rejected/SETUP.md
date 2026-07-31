# Scenario

**Feature**: unrelated exclusive modes remain invalid with unwind

```
--unwind plus --list -> validation error
```

```go
import "github.com/xhd2015/doctest/session"
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; req.Args=[]string{"--unwind","--list"}; return nil }
```
