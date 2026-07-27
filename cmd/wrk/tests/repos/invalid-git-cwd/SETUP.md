# Scenario

**Feature**: wrk --repos rejects cwd outside any git repository

```
plain cwd -> wrk --repos -> error (not a git repository)
```

## Steps

- Descendant scenarios run `wrk --repos` from a plain directory.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--repos"}
	return nil
}
```
