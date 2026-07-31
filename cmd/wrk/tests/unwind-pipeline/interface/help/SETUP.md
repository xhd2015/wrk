# Scenario

**Feature**: help documents unwind partners

```
--help -> CLI usage -> sync and generated-commit unwind partners
```

```go
import "github.com/xhd2015/doctest/session"
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; req.Args=[]string{"--help"}; return nil }
```
