# Scenario

**Feature**: wrk --all-deps --dry-run with empty projects prints would: wrked 0 deps

```
# no projects.json -> would: wrked 0 deps, no side effects
consumer (requires dep1) + empty projects -> wrk --all-deps --dry-run
```

## Steps

1. Create a consumer requiring `example.com/dep1`.
2. Do not register any projects.
3. Run `wrk --all-deps --dry-run` from the consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	allDepsEnsureHelpersUsed()
	dryRunEnsureHelpersUsed()

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	req.InProcess = true
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--dry-run"}
	return nil
}
```