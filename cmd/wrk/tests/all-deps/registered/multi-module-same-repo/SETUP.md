# Scenario

**Feature**: wrk --all-deps dedups to one worktree when multiple required modules share a registered repo

```
# registered myrepo has svc-a=dep1 and svc-b=dep2; consumer requires both -> one worktree, two replaces
projects.json (myrepo) + consumer requires dep1+dep2 -> wrked 2 deps
```

## Steps

1. Create and register `myrepo` with two sub-modules `svc-a` and `svc-b`.
2. Create a consumer requiring both modules.
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()
	registeredEnsureHelpersUsed()

	myrepo := allDepsDepDir(req.WorkRoot, "myrepo")
	initMultiModuleRepo(t, myrepo)
	registerAllDepsProject(t, req, myrepo)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```