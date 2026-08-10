# Scenario

**Feature**: two sequential cascade pins into same dirty consumer (P3-3)

```
# root requires shared + other (local replaces); both free modules owned-changed
# dirty go.mod WIP; no --add-all
root <- shared, root <- other
  -> wrk --unwind --tag-next --push
  -> tag free modules free-first; pin root twice (or chained commits)
  -> both require bumps on pin commit tree AND restored WT (+ WIP marker)
```

## Steps

1. Seed three-module apply fixture (`setupApplyCascadeSingleRepoThreeModules`).
2. Dirty root go.mod WIP.
3. Run without `--add-all`.
4. Assert both surgical bumps on WT and in pin commit history.

## Context

- Free-first: both leaf modules tag before consumer pins; pin order stable by dep path.
- Partial edit must chain across multiple pins without dropping the first bump or WIP.
- **RED** until partial edit + multi-pin restore.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeSingleRepoThreeModules(t, req)
	dirtyRootGoModWIP(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push"}
	return nil
}
```
