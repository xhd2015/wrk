# Scenario

**Feature**: acyclic `--unwind --dry-run` plan paths (order, skip, flags, gen-commit)

```
# residual DAG among dirty pending → free-first would: peel <display-path>
# missing pin flags with edges → hard Error before plan apply
# gen-commit subtree: add-all / leave-N vocabulary
stack (acyclic) -> wrk --unwind --dry-run [flags] -> plan or flag error
```

## Preconditions

- Parent `unwind/` helpers: `setupThreeRepoChain`, `setupSingleMainDirty`,
  `dirtyAllThree`, `dirtyMidAndRoot`, `setPeelOrderDisplays`, baseline asserts.
- Leaves set `req.InProcess = true` and full `req.Args` (including `--unwind`).
- PeelOrder holds **display paths** (external/…, `.`), not bare MainRepo basenames.

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
