# Scenario

**Feature**: wrk --dep errors when dep path is not a git repository

```
# plain directory without .git -> wrk --dep -> non-zero
```

## Steps

1. Create consumer with dep require.
2. Create non-git directory as dep path.
3. Run `wrk --dep`.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	dep := filepath.Join(req.WorkRoot, "not-git")
	mkdirAll(t, dep)
	writeFile(t, filepath.Join(dep, "go.mod"), "module "+depModulePath+"\n\ngo 1.22\n")

	req.RepoDir = consumer
	req.DepPath = dep
	req.Args = []string{"--dep", dep}
	return nil
}
```