# Scenario

**Feature**: --json is not valid with --propagate-tags (even when --tag-next is set)

```
workspace/ -> wrk --tag-next --propagate-tags --json
  -> non-zero
  -> stderr names --json and --propagate-tags
  -> no JSON plan on stdout
```

## Steps

1. Neutral cwd (flag-layer reject; full compose fixtures not required).
2. Run compose with `--json`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--tag-next", "--propagate-tags", "--json"}
	return nil
}
```
