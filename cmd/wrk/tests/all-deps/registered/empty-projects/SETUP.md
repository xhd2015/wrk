# Scenario

**Feature**: wrk --all-deps with empty projects.json makes no changes

```
# no projects.json (or empty list) -> wrked 0 deps, no external/, no replaces
consumer (requires dep1) + empty projects.json -> wrk --all-deps -> wrked 0 deps
```

## Steps

1. Create a consumer requiring `example.com/dep1`.
2. Do not register any projects (`projects.json` absent).
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```