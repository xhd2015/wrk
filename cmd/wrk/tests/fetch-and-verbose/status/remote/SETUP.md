# Scenario

**Feature**: wrk --status adds Remote: to root Dir: . block on main checkout cwd

```
main repo cwd -> root block gains Remote: (same brief labels as --projects)
linked wt cwd -> no Remote: on any block
nested RepoTypeMain -> no Remote:
```

## Steps

- Descendants run `wrk --status` and assert `Remote:` presence rules.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureFetchVerboseHelpersUsed()
	return nil
}
```