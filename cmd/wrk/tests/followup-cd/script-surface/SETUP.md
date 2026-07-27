# Scenario

**Feature**: bash integration script surface for follow-up auto-cd

```
# print / install / complete expose wrapper + --no-cd
wrk --bash-integration [...] -> script or flag list includes follow-up surface
```

## Steps

1. Group script-surface leaves; descendants set Mode print|install|complete.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// script-surface does not require a git checkout
	if req.RepoDir == req.WorkRoot || req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	_ = scriptDefinesWrkWrapper
	return nil
}

func scriptDefinesWrkWrapper(script string) bool {
	for _, line := range strings.Split(script, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "wrk()") || strings.HasPrefix(trim, "function wrk") {
			return true
		}
	}
	return false
}
```
