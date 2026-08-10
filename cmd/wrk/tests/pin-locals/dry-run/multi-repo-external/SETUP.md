# Scenario

**Feature**: dry-run plans relative pin for nested external dep already required

```
primary requires example.com/dep; external/dep in stack
  -> wrk --pin-locals --dry-run
  -> would: pin-local example.com/consumer <- example.com/dep => ./external/dep
     (or ../external/… form if rel differs; asserts accept either relative form)
  -> go.mod unchanged
  -> exit 0
```

## Steps

1. Build multi-repo external consumer fixture (no replace yet).
2. Run dry-run.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMultiRepoExternalConsumer(t, req)
	req.Args = []string{"--pin-locals", "--dry-run"}
	return nil
}
```
