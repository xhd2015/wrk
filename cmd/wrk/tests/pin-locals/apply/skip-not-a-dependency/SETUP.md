# Scenario

**Feature**: apply does not add replace for inventory module that is not a dependency

```
primary requires dep; external/other not required
  -> wrk --pin-locals
  -> may pin dep
  -> must NOT add replace for example.com/other
  -> exit 0
```

## Steps

1. Build skip-not-a-dependency fixture.
2. Run apply.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSkipNotADependency(t, req)
	req.Args = []string{"--pin-locals"}
	return nil
}
```
