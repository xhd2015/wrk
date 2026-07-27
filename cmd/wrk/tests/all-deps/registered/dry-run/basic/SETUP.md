# Scenario

**Feature**: wrk --all-deps --dry-run prints would: lines for registered deps without writing

```
# registered mydep1+mydep2 -> would: lines for both, no external/, no replaces
projects.json (mydep1, mydep2) + consumer -> wrk --all-deps --dry-run -> would: wrked 2 deps
```

## Steps

1. Create and register `mydep1` and `mydep2`.
2. Create a consumer requiring both modules.
3. Run `wrk --all-deps --dry-run` from the consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()
	dryRunEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	initAllDepsRepo(t, dep2, "example.com/dep2", "dep2")
	registerAllDepsProjects(t, req, dep1, dep2)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--dry-run"}
	return nil
}
```