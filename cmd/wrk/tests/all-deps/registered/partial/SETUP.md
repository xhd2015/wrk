# Scenario

**Feature**: wrk --all-deps links only registered deps that match required modules

```
# consumer requires dep1+dep2; only dep1 registered -> dep1 linked, dep2 untouched
consumer (requires dep1, dep2) + projects.json (mydep1 only) -> wrk --all-deps -> wrked 1 deps
```

## Steps

1. Create and register only `mydep1` (`example.com/dep1`).
2. Create a consumer requiring `example.com/dep1` and `example.com/dep2`.
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	registerAllDepsProject(t, req, dep1)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```