# Scenario

**Feature**: `--done` peel removes leaf linked worktree under external/

```
# same 2-repo apply stack as leaf-then-pin
root main + leaf ext (dirty+ahead)
  -> wrk --unwind --done --tag-next --push
  -> peel leaf with --done → path under external/ gone
  -> pin still applied on root main go.mod
```

## Steps

1. Build 2-repo apply stack.
2. Run unwind with `--done` + pin flags from root main.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyLeafPinStack(t, req)
	req.Args = []string{"--unwind", "--done", "--tag-next", "--push"}
	return nil
}
```
