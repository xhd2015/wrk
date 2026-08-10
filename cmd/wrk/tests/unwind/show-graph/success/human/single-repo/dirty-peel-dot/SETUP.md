# Scenario

**Feature**: dirty single main — peel includes `.`; no pin flags required

```
root (dirty, main) -> wrk --unwind --show-graph
  -> peel order includes .
  -> exit 0 without --tag-next/--push
```

## Steps

1. Seed dirty single main.
2. Run show-graph only.
3. PeelOrder = `.`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
