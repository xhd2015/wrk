# Scenario

**Feature**: wrk --dep creates external worktree, replace, tidy, and gitignore entry

```
# consumer requires dep -> wrk --dep dep-repo -> external/mydep-main-{date} + replace + /external gitignore
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Create dep git repo `mydep` with module `example.com/dep`.
3. Run `wrk --dep <dep>` from consumer.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	dep := initDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", dep}
	return nil
}
```