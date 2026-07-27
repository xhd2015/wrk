# Scenario

**Feature**: tracked up-to-date main repo shows Remote: identical on root block

```
tracked main up-to-date -> wrk --status -> root Remote: identical
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```