# Scenario

**Feature**: wrk skill --show <unknown> fails and lists available skills

```
workspace/ -> wrk skill --show nope -> exit ≠0, stderr unknown skill
```

## Steps

1. Run `wrk skill --show nope` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show", "nope"}
	return nil
}
```
