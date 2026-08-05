# Scenario

**Feature**: no local project with matching github origin → error

```
unrelated non-github project only; neutral cwd
  -> wrk --where --pr https://github.com/acme/app/pull/42
  -> non-zero; empty stdout; stderr mentions no local project / owner/repo
```

## Steps

1. Record a non-github origin project only.
2. Fake gh succeeds with headRefName.
3. Run from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	wherePrSetupNoMatchingProject(t, req)
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```
