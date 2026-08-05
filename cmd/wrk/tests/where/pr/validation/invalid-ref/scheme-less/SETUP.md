# Scenario

**Feature**: scheme-less path / host forms are not accepted

```
workspace/ -> wrk --where --pr github.com/acme/app/pull/42
  -> non-zero; empty stdout
  -> stderr: full GitHub pull request URL required
```

## Steps

1. Neutral cwd.
2. Args: `--where --pr github.com/acme/app/pull/42` (no `https://`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where", "--pr", "github.com/acme/app/pull/42"}
	return nil
}
```
