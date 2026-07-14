# Scenario

**Feature**: wrk --dry-run without --all-deps is rejected before any planning

```
# bare --dry-run (no --all-deps) -> non-zero exit, stderr mentions --dry-run is only valid with --all-deps
wrk --dry-run -> error (--dry-run is only valid with --all-deps)
```

## Steps

1. Create a consumer git repo requiring `example.com/dep1` (so cwd is a valid git repo with a go.mod, though the error fires before scanning regardless).
2. Run `wrk --dry-run` from the consumer (no `--all-deps`).

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--dry-run"}
	return nil
}
```
