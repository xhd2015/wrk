# Scenario

**Feature**: wrk --all-deps tolerates and skips already-replaced modules

```
# dep1 pre-replaced; dep2 registered -> dep1 skipped, dep2 linked
consumer (dep1 pre-replaced) + projects.json (mydep1, mydep2) -> wrked 1 deps
```

## Steps

1. Create and register `mydep1` and `mydep2`.
2. Create a consumer requiring both, with `replace example.com/dep1 => ./external/preexisting`.
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	initAllDepsRepo(t, dep2, "example.com/dep2", "dep2")
	registerAllDepsProjects(t, req, dep1, dep2)

	extraGoMod := "replace example.com/dep1 => ./external/preexisting"
	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, extraGoMod)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```