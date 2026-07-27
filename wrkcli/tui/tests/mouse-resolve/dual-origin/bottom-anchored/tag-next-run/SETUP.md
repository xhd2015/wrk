# Scenario

**Feature**: bottom-anchored tag-next Run resolves correctly

```
# bottom-anchored dual-origin
absY = origin + tag-next localY
  -> runStage == "tag-next"
```

## Steps

1. Aim Run chip for `tag-next`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.StageID = "tag-next"
	req.Target = "run"
	return nil
}
```
