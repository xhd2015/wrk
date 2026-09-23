# Scenario

**Feature**: wrk -h documents the --install flag

```
# I17: wrk -h
  -> exit 0; help text contains --install (with the one-name minimum)
```

## Steps

1. Run `wrk -h` from the empty fixture module root (no planning happens).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```
