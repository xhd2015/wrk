# Scenario

**Feature**: default wrk --projects skips fetch; stale tracking ref shows identical

```
push to origin without local fetch -> Remote: identical
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}
```