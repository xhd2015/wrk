# Scenario

**Feature**: summary hint lists land/pin flags apply would need (not an error)

```
# linked dirty multi-repo with edges; no apply flags passed
wrk --unwind --show-graph
  -> exit 0 (do not require --tag-next/--push/--merge-back)
  -> summary hint: apply would need --merge-back --tag-next --push
  -> zero mutations
```

## Steps

1. Build three-repo dirty chain (linked consumer + edges).
2. Run show-graph without any land/pin flags.
3. Assert optional hint — not a hard error.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyAllThree(t, req)
	setPeelOrderDisplays(t, req, req.DepsLinkedWtDir, req.ExternalWtDir, req.WtDir)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
