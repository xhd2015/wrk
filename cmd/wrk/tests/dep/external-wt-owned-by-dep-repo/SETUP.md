# Scenario

**Feature**: wrk --dep spawns the external worktree as a worktree of the DEP repo (registered under the dep's main repo, not the consumer's)

```
# consumer requires dep -> wrk --dep dep-repo -> external/mydep-main-{date}
# external worktree's .git gitdir must point into <dep-main>/.git/worktrees/, NOT <consumer>/.git/worktrees/
```

## Preconditions

- Git and Go must be available.
- Consumer cwd must be inside a git work tree with a `go.mod`.
- Dep path must be a git repo with a valid Go module listed in consumer go.mod.

## Steps

1. Create consumer git repo (main checkout) with `go.mod` requiring `example.com/dep`.
2. Create dep git repo `mydep` (main checkout) with module `example.com/dep`.
3. Run `wrk --dep <dep>` from the consumer.

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
