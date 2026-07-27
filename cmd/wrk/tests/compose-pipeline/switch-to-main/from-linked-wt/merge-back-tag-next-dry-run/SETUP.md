# Scenario

**Feature**: `--merge-back --tag-next` from linked wt: merge keeps WT; tag-next runs on main after activeRoot switch

```
# Linked wt ahead of v0.0.1
linked wt
  -> wrk --merge-back --tag-next --dry-run
  -> merge --ff-only plan WITHOUT worktree remove / branch -D
  -> activeRoot := main → tag-next plans v0.0.2 on main tip
  -> wt remains; no v0.0.2 created
```

## Steps

1. Linked ahead fixture; baseline.
2. Run `--merge-back --tag-next --dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAPLinkedAhead(t, req)
	recordAPDryRunBaseline(t, req)
	req.Args = []string{"--merge-back", "--tag-next", "--dry-run"}
	return nil
}
```
