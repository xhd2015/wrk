# Scenario

**Feature**: flag order free for `--where --pr` compose

```
--pr --where URL  and  URL --where --pr
  -> same path stdout as --where --pr URL
```

## Preconditions

- Same fixture as single projects-json happy path.

## Steps

- Leaves only change argv order / positional placement.

## Context

- Behavioral equivalence to flag-first form.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	return nil
}
```
