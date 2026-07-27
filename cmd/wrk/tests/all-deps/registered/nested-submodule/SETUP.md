# Scenario

**Feature**: wrk --all-deps links a required nested sub-module from a registered project

```
# registered myrepo has root module + nested services/dep; consumer requires sub-module only
projects.json (myrepo) + consumer requires example.com/dep -> one worktree + replace at sub-dir
```

## Steps

1. Create and register `myrepo` with nested `services/dep` module `example.com/dep`.
2. Create a consumer requiring `example.com/dep`.
3. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()
	registeredEnsureHelpersUsed()

	myrepo := allDepsDepDir(req.WorkRoot, "myrepo")
	initNestedSubmoduleRepo(t, myrepo)
	registerAllDepsProject(t, req, myrepo)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```