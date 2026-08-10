# Scenario

**Feature**: require path not owned by any stack module is not pinned

```
primary requires example.com/missing; stack has example.com/dep (not required)
  -> wrk --pin-locals
  -> no pin for missing
  -> no inventing pin of dep
  -> already up to date / applied 0
  -> exit 0
```

## Steps

1. Build skip-no-matching-owner fixture.
2. Run apply.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSkipNoMatchingOwner(t, req)
	req.Args = []string{"--pin-locals"}
	return nil
}
```
