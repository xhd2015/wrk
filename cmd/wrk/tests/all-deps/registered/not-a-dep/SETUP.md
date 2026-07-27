# Scenario

**Feature**: wrk --all-deps skips registered projects whose modules are not required by the consumer

```
# registered otherrepo (example.com/other) not in consumer requires -> skipped
projects.json (otherrepo) + consumer requires dep1 only -> wrked 0 deps
```

## Steps

1. Create and register `otherrepo` with module `example.com/other`.
2. Create a consumer requiring only `example.com/dep1` (no matching registered project).
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()

	other := allDepsDepDir(req.WorkRoot, "otherrepo")
	initAllDepsRepo(t, other, "example.com/other", "other")
	registerAllDepsProject(t, req, other)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```