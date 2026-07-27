# Scenario

**Feature**: two-arg second positional as single component >255 bytes, non-TTY → error + hint

```
wrk <myrepo> <256 x 'b'> (non-TTY)
  -> task-like by ENAMETOOLONG / component length class
  -> Error + -t hint (not raw ENAMETOOLONG-only without task hint)
```

## Steps

1. Init `myrepo`.
2. Second positional = 256 ASCII `b` bytes (no whitespace, no `/`).

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoArg(t, req, strings.Repeat("b", 256))
	return nil
}
```
