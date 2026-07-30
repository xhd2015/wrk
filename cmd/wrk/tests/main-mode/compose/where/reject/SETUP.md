# Scenario

**Feature**: wrk --main --where rejection paths

```
wrk --main --where foo        -> unexpected arguments
wrk --main --where (not git)  -> not a git repository
wrk --main --where --exec …   -> --exec not valid with --where
wrk --main --where --list     -> mutually exclusive
```

## Steps

- Descendants set invalid Args and/or non-git cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	return nil
}
```
