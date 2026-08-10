# Scenario

**Feature**: apply rewrites absolute replace to relative same-target

```
primary has replace example.com/dep => /abs/.../external/dep
  -> wrk --pin-locals
  -> pin-local rewrite line
  -> go.mod NewPath relative (./external/dep)
  -> exit 0
```

## Steps

1. Build rewrite-absolute fixture.
2. Run apply.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupRewriteAbsolute(t, req)
	req.Args = []string{"--pin-locals"}
	return nil
}
```
