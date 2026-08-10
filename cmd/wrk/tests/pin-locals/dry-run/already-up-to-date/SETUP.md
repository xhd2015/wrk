# Scenario

**Feature**: dry-run when relative replace already correct reports already up to date

```
primary has replace example.com/dep => ./external/dep
  -> wrk --pin-locals --dry-run
  -> already up to date style message
  -> no would: pin-local apply lines for work
  -> go.mod unchanged
  -> exit 0
```

## Steps

1. Build already-relative fixture.
2. Run dry-run.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAlreadyRelativePin(t, req)
	req.Args = []string{"--pin-locals", "--dry-run"}
	return nil
}
```
