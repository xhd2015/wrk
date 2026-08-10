# Scenario

**Feature**: apply writes relative replace for nested external dep

```
primary requires example.com/dep; external/dep in stack; no replace
  -> wrk --pin-locals
  -> pin-local example.com/consumer <- example.com/dep => ./external/dep
  -> go.mod has relative replace (not absolute)
  -> summary applied >= 1, tidy ok >= 1
  -> exit 0
```

## Steps

1. Build multi-repo external consumer fixture.
2. Run apply.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMultiRepoExternalConsumer(t, req)
	req.Args = []string{"--pin-locals"}
	return nil
}
```
