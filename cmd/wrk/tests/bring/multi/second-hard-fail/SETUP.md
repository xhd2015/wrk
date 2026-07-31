# Scenario

**Feature**: multi-bring fail-fast when the second path is a hard error (missing)

```
# first dep succeeds; second path does not exist
# wrk --bring <dep1> --bring <missing>
#   -> non-zero; first external kept; second not created
#   -> stdout may include first path (partial); stop before dep2 worktree
consumer + mydep1 + missing path
  -> multi-bring -> fail-fast hard error
```

## Steps

1. Create consumer requiring dep1+dep2 modules and a valid `mydep1` repo.
2. Point second `--bring` at a path that does not exist under WorkRoot.
3. Run multi-bring (L2).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initMultiBringConsumerWithTwoRequires(t, req.WorkRoot)
	dep1 := initMultiBringDepRepo(t, req.WorkRoot, "mydep1", multiBringDep1Module)
	// Hard-missing second path (never created).
	missing := filepath.Join(req.WorkRoot, "missing-dep-does-not-exist")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.SecondRepo = missing
	req.Args = []string{"--bring", dep1, "--bring", missing}
	return nil
}
```
