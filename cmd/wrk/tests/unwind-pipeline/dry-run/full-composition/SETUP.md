# Scenario

**Feature**: complete dry-run composition is accepted

```
linked dependency -> all unwind modifiers -> planned commit through reinstall
```

```go
import "github.com/xhd2015/doctest/session"
func Setup(t *testing.T, d *session.Doctest, req *Request) error { _ = d; req.Args=[]string{"--unwind","--gen-commit-msg","--commit","--agent-runner=commandcode","--merge-back","--sync","--tag-next","--push","--reinstall-local","--dry-run"}; return nil }
```
