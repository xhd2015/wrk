# Scenario

**Feature**: two-arg second positional >120 bytes (no spaces), non-TTY → default auto-promote

```
wrk <myrepo> <121 x 'a'> (non-TTY)
  -> task-like by length; auto-promote under WRK_HOME with fitted slug
```

## Steps

1. Init `myrepo`.
2. Second positional = 121 ASCII `a` bytes (no whitespace).

```go
import (
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	arg := strings.Repeat("a", 121)
	setupTwoArg(t, req, arg)
	return nil
}
```
