# Scenario

**Feature**: wrk --dep errors when dep module is not in consumer go.mod

```
# consumer without dep require -> wrk --dep -> non-zero
```

## Steps

1. Create consumer **without** requiring dep.
2. Create valid dep repo.
3. Run `wrk --dep`.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerRepo(t, req.WorkRoot, false)
	dep := initDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.Args = []string{"--dep", dep}
	return nil
}
```