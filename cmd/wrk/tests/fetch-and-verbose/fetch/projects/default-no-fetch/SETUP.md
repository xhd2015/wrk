# Scenario

**Feature**: default wrk --projects skips fetch; stale tracking ref shows identical

```
push to origin without local fetch -> Remote: identical
```

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	return nil
}
```