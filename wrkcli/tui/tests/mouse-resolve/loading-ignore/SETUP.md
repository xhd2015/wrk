# Scenario

**Feature**: loading gate ignores mouse clicks

```
# loading resolve
Loading=true, click on any Run chip
  -> ResolveMouseHit
  -> miss (OK == false)
```

## Preconditions

- Uses dual-origin top-anchored geometry with a real Run aim so the only reason
  for miss is the loading flag.

## Steps

1. Set Loading true for descendant leaves.
2. Leave origin unknown; use simple top-anchored absY.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "resolve"
	req.Loading = true
	req.OriginYSet = false
	req.ExtraBlank = 10
	return nil
}
```
