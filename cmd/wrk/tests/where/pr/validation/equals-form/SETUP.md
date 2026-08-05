# Scenario

**Feature**: `--pr=URL` equals form is rejected (Bool flag)

```
workspace/ -> wrk --where --pr=https://github.com/acme/app/pull/42
  -> non-zero; empty stdout
  -> stderr mentions --pr / equals / invalid form
```

## Steps

1. Neutral cwd.
2. Args: `--where` and `--pr=<url>` equals form (no separate positional).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where", "--pr=" + wherePrURL}
	return nil
}
```
