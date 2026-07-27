# Scenario

**Feature**: wrk --bring soft-skips replace when dep module is not required by consumer

```
# consumer has go.mod but no require of dep -> wrk --bring
#   -> exit 0; external worktree + /external gitignore
#   -> no replace; SKIP not a dependency on stderr
consumer (go.mod, no require) + mydep (module example.com/dep)
  -> wrk --bring <dep>
  -> stdout external path; stderr SKIP … not a dependency of any consumer module
```

## Steps

1. Create consumer **without** requiring dep.
2. Create valid dep repo with go.mod.
3. Run `wrk --bring <dep>` from consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, false)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", dep}
	return nil
}
```
