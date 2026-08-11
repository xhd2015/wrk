# Scenario

**Feature**: --all soft-skips inventory-owned require when owner has no tags

```
# registered lib with no version tags; app requires example.com/lib@v0.0.1
cwd=app -> wrk --dep-update --all
  -> warning: on stderr (no tag / no version)
  -> exit 0
  -> no pin line; summary updated 0, skipped ≥1
  -> app go.mod unchanged
```

## Steps

1. Seed untagged owner + app require; register owner.
2. Run apply from app.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllNoTagOwner(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```
