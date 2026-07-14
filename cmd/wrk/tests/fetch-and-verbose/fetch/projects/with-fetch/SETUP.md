# Scenario

**Feature**: wrk --projects --fetch refreshes upstream before Remote: comparison

```
--fetch runs git fetch before CompareBranches -> accurate Remote: label
```

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	return nil
}
```