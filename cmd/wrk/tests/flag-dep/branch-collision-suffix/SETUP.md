# Scenario

**Feature**: external dep preferred branch pre-exists → joint path+branch `-1` (P2 collision; always new branch)

```
# dep main has refs/heads/main-{date}; path external/mydep-main-{date} free
consumer --dep mydep
  -> path external/mydep-main-{date}-1
  -> branch main-{date}-1 (no mydep- basename; -b new branch)
```

## Steps

1. Create consumer requiring `example.com/dep`.
2. Create dep repo `mydep` on `main`.
3. Pre-create branch `main-{WRK_DATE}` in the dep repo.
4. Run `wrk --dep <dep>` from consumer.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	dep := initDepRepo(t, req.WorkRoot, "mydep", true)
	runGitIsolated(t, dep, "branch", branchName("main", wrkDate, 0))

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", dep}
	return nil
}
```
