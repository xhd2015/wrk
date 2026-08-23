# Scenario

**Feature**: wrk --bring creates external worktree, replace, tidy, and gitignore when dep is required

```
# consumer requires dep -> wrk --bring dep-repo
#   -> external/mydep + replace + /external gitignore
#   -> stdout abs path; no SKIP on stderr
consumer (go.mod require example.com/dep) + mydep (module example.com/dep)
  -> wrk --bring <dep> -> success with replace
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Create dep git repo `mydep` with module `example.com/dep`.
3. Run `wrk --bring <dep>` from consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", dep}
	return nil
}
```
