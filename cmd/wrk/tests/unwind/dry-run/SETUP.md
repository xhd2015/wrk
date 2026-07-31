# Scenario

**Feature**: acyclic `--unwind --dry-run` plan paths (order, skip, flags)

```
# residual DAG among dirty pending → free-first would: peel lines
# missing pin flags with edges → hard Error before plan apply
stack (acyclic) -> wrk --unwind --dry-run [flags] -> plan or flag error
```

## Preconditions

- Parent `unwind/` helpers: `setupThreeRepoChain`, `setupSingleMainDirty`,
  `dirtyAllThree`, `dirtyMidAndRoot`, baseline zero-mutation asserts.
- Leaves set `req.InProcess = true` and full `req.Args` (including `--unwind`).

## Steps

1. Grouping marks the dry-run acyclic family; leaves seed fixtures and flags.
2. All leaves under this node plan or validate under `--dry-run` (no apply).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Acyclic dry-run family: every leaf includes --dry-run (leaves may prepend
	// --unwind and ship/land flags). Keep helpers referenced for generator.
	_ = t
	req.InProcess = true
	unwindEnsureHelpersUsed()
	return nil
}
```
