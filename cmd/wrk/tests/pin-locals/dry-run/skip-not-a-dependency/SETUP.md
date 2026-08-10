# Scenario

**Feature**: dry-run does not plan pin for inventory module that is not a dependency

```
primary requires dep; external/other present but not required
  -> wrk --pin-locals --dry-run
  -> may would: pin dep
  -> must NOT would: pin for example.com/other
  -> go.mod unchanged
```

## Steps

1. Build skip-not-a-dependency fixture.
2. Run dry-run.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSkipNotADependency(t, req)
	req.Args = []string{"--pin-locals", "--dry-run"}
	return nil
}
```
