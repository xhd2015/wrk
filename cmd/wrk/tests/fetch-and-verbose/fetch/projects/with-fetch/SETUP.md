# Scenario

**Feature**: wrk --projects --fetch refreshes upstream before Remote: comparison

```
--fetch runs git fetch before CompareBranches -> accurate Remote: label
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}
```