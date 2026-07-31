# Scenario

**Feature**: peel dirty leaf WT free-first then Pin consumer require to new tag

```
# root (main) requires example.com/dot-pkgs@v0.0.1
# leaf linked WT under root/external: ahead + DIRTY; bare origin on leaf main
root main + leaf ext
  -> wrk --unwind --done --tag-next --push
  -> peel dot-pkgs: land → tag v0.0.2 → push bare
  -> Pin root go.mod require → v0.0.2 (no replace); tidy
  -> leaf main advanced; origin has tag+main
```

## Steps

1. Build 2-repo apply stack (`setupApplyLeafPinStack`).
2. Run non-dry-run unwind with land + pin flags from root main.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyLeafPinStack(t, req)
	// Flag order free; --done lands linked leaf; pin via tag-next+push.
	req.Args = []string{"--unwind", "--done", "--tag-next", "--push"}
	return nil
}
```
