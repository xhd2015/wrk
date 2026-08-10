# Scenario

**Feature**: A↔B cycle dry-run rejects before cascade plan (C-DR5)

```
A ↔ B -> wrk --unwind --dry-run --tag-next --push --done
  -> cycle error; no successful cascade body; zero mutations
```

## Steps

1. Build two-repo cycle stack (`setupTwoCycleStack`).
2. Run dry-run with pin + land flags (same as cycle/two-cycle).
3. Assert cycle error + no cascade tag plan.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoCycleStack(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```
