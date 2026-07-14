# Scenario

**Feature**: wrk --all-deps links every required dependency registered in projects.json

```
# consumer requires dep1+dep2; both registered -> both linked in project-path order
consumer (requires dep1, dep2) + projects.json (mydep1, mydep2) -> wrk --all-deps -> 2 external wts + wrked 2 deps
```

## Steps

1. Create dep repos `mydep1` (`example.com/dep1`) and `mydep2` (`example.com/dep2`) under `{WorkRoot}/deps/`.
2. Register both via `wrk --add`.
3. Create a consumer requiring both modules.
4. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()
	registeredEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	initAllDepsRepo(t, dep2, "example.com/dep2", "dep2")
	registerAllDepsProjects(t, req, dep1, dep2)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```