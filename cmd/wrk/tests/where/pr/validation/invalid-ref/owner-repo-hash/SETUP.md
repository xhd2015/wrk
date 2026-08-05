# Scenario

**Feature**: `owner/repo#N` shorthand is not accepted

```
workspace/ -> wrk --where --pr acme/app#42
  -> non-zero; empty stdout
  -> stderr: full GitHub pull request URL required
```

## Steps

1. Neutral cwd.
2. Args: `--where --pr acme/app#42`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where", "--pr", "acme/app#42"}
	return nil
}
```
