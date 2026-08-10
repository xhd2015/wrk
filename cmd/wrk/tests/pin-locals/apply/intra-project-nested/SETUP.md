# Scenario

**Feature**: apply pins intra-project nested module with ./tools

```
root requires example.com/root/tools; tools/ nested go.mod
  -> wrk --pin-locals
  -> pin-local example.com/root <- example.com/root/tools => ./tools
  -> relative replace present
  -> exit 0
```

## Steps

1. Build intra-project nested fixture.
2. Run apply.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupIntraProjectNested(t, req)
	req.Args = []string{"--pin-locals"}
	return nil
}
```
