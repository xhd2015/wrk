# Scenario

**Feature**: unwind preserves partial-success semantics and stops hard failures

```
peeled repository -> sync warning or hard release failure -> continue or stop as specified
```

```go
import "github.com/xhd2015/doctest/session"
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=t; _=d; _=req; return nil }
```
