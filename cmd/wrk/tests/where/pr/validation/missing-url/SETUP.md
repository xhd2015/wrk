# Scenario

**Feature**: `--where --pr` with zero positionals → requires full GitHub PR URL

```
workspace/ -> wrk --where --pr
  -> non-zero; empty stdout
  -> stderr requires full GitHub pull request URL (not basename-only wording alone)
```

## Steps

1. Neutral cwd; no positionals.
2. Args: `--where --pr` only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where", "--pr"}
	return nil
}
```
