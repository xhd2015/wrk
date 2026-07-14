# Scenario

**Feature**: wrk --all-deps skips the consumer's own main repo when it appears in projects.json

```
# consumer and mydep1 both registered; consumer requires dep1 -> dep1 linked, self skipped
projects.json (consumer, mydep1) + consumer requires dep1 -> wrked 1 deps
```

## Steps

1. Create `mydep1` (`example.com/dep1`) and register it.
2. Create the consumer requiring `example.com/dep1` and register the consumer main repo too.
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")
	registerAllDepsProjects(t, req, consumer, dep1)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```