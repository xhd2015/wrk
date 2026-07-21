# Scenario

**Feature**: wrk --dep mydep creates external worktree from single saved dep

```
saved/mydep in projects.json (module example.com/dep)
consumer requires dep, no ./mydep -> wrk --dep mydep -> external/mydep-main-{date}
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Create dep git repo at `{WorkRoot}/saved/mydep` with module `example.com/dep`.
3. Record saved dep with `wrk --add`.
4. Run `wrk --dep mydep` from consumer cwd.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerForDepBasename(t, req.WorkRoot)
	savedDep := initSavedDepRepo(t, req.WorkRoot, "saved", "mydep")
	recordSavedProject(t, req, savedDep)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = savedDep
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", "mydep"}
	return nil
}
```