# Scenario

**Feature**: wrk --all-deps is mutually exclusive with --dep

```
# wrk --all-deps --dep <dep> -> non-zero exit, stderr mentions mutually exclusive
wrk --all-deps --dep <dep> -> error (mutually exclusive)
```

## Steps

1. Create a consumer git repo requiring `example.com/dep1`.
2. Create a dep repo `mydep1` (module `example.com/dep1`).
3. Run `wrk --all-deps --dep <dep>` from the consumer (no `--scan-root`).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")
	dep := filepath.Join(req.WorkRoot, "mydep1")
	initAllDepsRepo(t, dep, "example.com/dep1", "dep1")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep
	req.Args = []string{"--all-deps", "--dep", dep}
	return nil
}
```
