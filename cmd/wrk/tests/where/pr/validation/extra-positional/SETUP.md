# Scenario

**Feature**: two+ positionals with `--where --pr` → unexpected arguments

```
workspace/ -> wrk --where --pr URL extra
  -> non-zero; empty stdout
  -> stderr: unexpected arguments
```

## Steps

1. Neutral cwd.
2. Args: `--where --pr <url> extra`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where", "--pr", wherePrURL, "extra"}
	return nil
}
```
