# Scenario

**Feature**: --fetch rejected when not combined with --projects or --status

```
wrk --fetch [other mode] -> exit 1, stderr --fetch is only valid with --projects or --status
```

## Steps

- Descendants set `req.Args` to invalid `--fetch` combinations.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureFetchVerboseHelpersUsed()
	req.RepoDir = req.WorkRoot
	return nil
}
```