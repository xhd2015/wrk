# Scenario

**Feature**: pin-when-primary-is-main — peel leaf then Pin primary Path (main is in scope)

```
# root primary *is* main (Path == MainRepo); requires example.com/dot-pkgs@v0.0.1
# leaf linked WT under root/external: ahead + DIRTY; bare origin on leaf main
root main + leaf ext
  -> wrk --unwind --done --tag-next --push  (RepoDir = root main)
  -> banner: ==== unwind: peel external/dot-pkgs-main-2026-06-30 ====
  -> peel leaf: land → tag v0.0.2 → push bare
  -> Pin root Path go.mod require → v0.0.2 (main is in scope; pin must edit it)
  -> leaf main advanced; origin has tag+main
```

## Steps

1. Build 2-repo apply stack (`setupApplyLeafPinStack`) — primary Path == MainRepo.
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
