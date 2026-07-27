# Scenario

**Feature**: wrk --status rejects cwd outside any git repository

```
# plain directory has no checkout root
plain cwd -> wrk --status -> error (not a git repository)
```

## Preconditions

- The effective cwd is not inside a git work tree.

## Steps

- Descendant scenarios run `wrk --status` from a plain directory.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--status"}
	return nil
}
```
