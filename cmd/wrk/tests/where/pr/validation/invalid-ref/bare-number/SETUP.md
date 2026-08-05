# Scenario

**Feature**: bare PR number is not accepted with `--where --pr`

```
workspace/ -> wrk --where --pr 390
  -> non-zero; empty stdout
  -> stderr: full GitHub pull request URL required
```

## Steps

1. Neutral cwd.
2. Args: `--where --pr 390`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where", "--pr", "390"}
	return nil
}
```
