# Scenario

**Feature**: replace target missing → `warning:` on stderr; plan continues for valid members

```
# root go.mod: replace => ../external/dot-pkgs-main-DATE (path does not exist)
# primary dirty; no other stack members required for plan
missing replace target
  -> wrk --unwind --dry-run
  -> stderr contains warning:
  -> would: peel .
  -> exit 0; zero mutations
```

## Steps

1. Seed consumer with local replace to a non-existent path.
2. Dirtify primary; dry-run (no pin flags — no successful cross-repo edge).
3. Assert warning + peel `.`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowMissingTarget(t, req)
	req.Args = []string{"--unwind", "--dry-run"}
	recordUnwindBaseline(t, req)
	return nil
}
```
