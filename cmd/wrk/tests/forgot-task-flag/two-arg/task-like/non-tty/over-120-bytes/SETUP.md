# Scenario

**Feature**: two-arg second positional >120 bytes (no spaces), non-TTY → error + hint

```
wrk <myrepo> <121 x 'a'> (non-TTY)
  -> task-like by length; Error + -t hint; no worktree
```

## Steps

1. Init `myrepo`.
2. Second positional = 121 ASCII `a` bytes (no whitespace).

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	arg := strings.Repeat("a", 121)
	setupTwoArg(t, req, arg)
	// Stash expected prose path rejection via SpawnDir value.
	return nil
}
```
