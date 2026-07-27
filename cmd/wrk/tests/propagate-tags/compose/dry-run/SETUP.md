# Scenario

**Feature**: compose dry-run plans tag-next and propagate using planned next tags

```
# same root-bump fixture as apply
cwd=lib -> wrk --tag-next --propagate-tags --dry-run
  -> (1) tag-next plan: v1.0.0 owned changed -> v1.0.1; "1 tag planned"
  -> blank line
  -> (2) propagate plan: source @ v1.0.1 (planned); would: update app v1.0.0 -> v1.0.1
  -> NO tag v1.0.1 created; app go.mod / HEAD / tags unchanged
  -> source go.mod / HEAD / tags unchanged
```

## Steps

1. `setupComposeRootBump` (no origin; proxy still ok, unused for dry-run).
2. Args: `--tag-next --propagate-tags --dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupComposeRootBump(t, req, false)
	req.Args = []string{"--tag-next", "--propagate-tags", "--dry-run"}
	return nil
}
```
