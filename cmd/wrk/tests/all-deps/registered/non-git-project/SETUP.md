# Scenario

**Feature**: wrk --all-deps silently skips registered non-git directories

```
# projects.json lists existing plain dir (not git) -> skip silently
consumer (requires dep1) + projects.json (non-git dir) -> wrked 0 deps
```

## Steps

1. Create a plain directory (no `.git`) and register it in `projects.json`.
2. Create a consumer requiring `example.com/dep1`.
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	nongit := filepath.Join(req.WorkRoot, "plain-dir")
	mkdirAll(t, nongit)
	writeFile(t, filepath.Join(nongit, "README.md"), "not a git repo\n")
	writeAllDepsProjectsJSON(t, req.WrkHome, allDepsResolvePath(t, nongit))

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```