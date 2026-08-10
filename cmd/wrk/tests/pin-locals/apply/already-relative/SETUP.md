# Scenario

**Feature**: apply when already correctly relative is a no-op

```
primary already has replace => ./external/dep
  -> wrk --pin-locals
  -> already up to date
  -> applied 0
  -> go.mod unchanged
  -> exit 0
```

## Steps

1. Build already-relative fixture.
2. Run apply.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAlreadyRelativePin(t, req)
	req.Args = []string{"--pin-locals"}
	return nil
}
```
