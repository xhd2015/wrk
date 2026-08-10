# Scenario

**Feature**: module-layer human polish (dir keys, replaced, drift)

```
# stack modules under checkouts
show-graph -> dir identity + collapsed → edges + replaced / (latest)
```

## Steps

1. Grouping scopes module-graph human leaves (follow-local-replace fixtures +
   require-drift).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```
