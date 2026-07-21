# Scenario

**Feature**: wrk --dep sub/mydep does not fall back to saved project basename mydep

```
saved/mydep recorded; consumer cwd -> wrk --dep sub/mydep -> does not exist (no fallback)
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Create and record saved dep at `{WorkRoot}/saved/mydep`.
3. Run `wrk --dep sub/mydep` from consumer cwd (no local `sub/mydep`).

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerForDepBasename(t, req.WorkRoot)
	savedDep := initSavedDepRepo(t, req.WorkRoot, "saved", "mydep")
	recordSavedProject(t, req, savedDep)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = savedDep
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", "sub/mydep"}
	return nil
}
```