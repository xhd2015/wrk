# Scenario

**Feature**: `--show-graph` cannot be used with `--dry-run`

```
wrk --unwind --show-graph --dry-run
  -> Error: show-graph … dry-run
  -> exit ≠ 0; no graph body
```

## Steps

1. Parent seeded single dirty main.
2. Run show-graph with `--dry-run`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--show-graph", "--dry-run"}
	return nil
}
```
